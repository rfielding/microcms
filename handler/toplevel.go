package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/rfielding/microcms/data"
	"github.com/rfielding/microcms/fs"
	"github.com/rfielding/microcms/utils"
)

var FileServer = http.FileServer(http.Dir(fs.At))

// Launch a plain http server
func Setup() {
	bindAddr := utils.Getenv("BIND", "0.0.0.0:9321")
	http.HandleFunc("/", rootRouter)
	log.Printf("start http at %s", bindAddr)
	log.Fatal(http.ListenAndServe(bindAddr, nil))
}

func postHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) {
	if len(pathTokens) > 2 && pathTokens[1] == "files" {
		postFilesHandler(w, r, pathTokens)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func detectNewUser(user string) (io.Reader, error) {
	pipeReader, pipeWriter := io.Pipe()
	go func() {
		userStruct := struct {
			Name string `json:"User"`
		}{Name: user}
		err := compiledDefaultPermissionsTemplate.Execute(pipeWriter, userStruct)
		if err != nil {
			panic(fmt.Sprintf("Unable to compile default rego template: %v", err))
		}
		pipeWriter.Close()
	}()
	return pipeReader, nil
}

func deleteHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) {
	if !CanWrite(data.GetUser(r), path.Dir(r.URL.Path), path.Base(r.URL.Path)) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(fmt.Sprintf("write disallowed")))
		return
	}
	if strings.HasPrefix(r.URL.Path, "/files/") {
		if fs.IsExist(r.URL.Path) {
			err := fs.Remove(r.URL.Path)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("unable to delete file: %v", err)))
				return
			}
		}
		// If it's gone, then it's deleted
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func ToBoolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func ensureThatHomeDirExists(w http.ResponseWriter, r *http.Request, user data.User) bool {
	// Make home directories exist on first visit
	if len(user["email"]) > 0 {
		userName := user["email"][0]
		parentDir := "/files/" + userName // i can't make it end in slash, seems inconsistent
		fsPath := parentDir + "/"
		fsName := "permissions.rego"
		if !fs.IsExist(fsPath + fsName) {
			log.Printf("Welcome to %s", fsPath)
			rdr, err := detectNewUser(userName)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				err2 := fmt.Errorf("Could not create homedir for %s: %v", userName, err)
				w.Write([]byte(fmt.Sprintf("%v", err2)))
				log.Printf("%v", err2)
				return true
			}
			if rdr != nil {
				herr, err := postFileHandler(user, rdr, fsPath, fsName, fsPath, fsName, false, true)
				if err != nil {
					w.WriteHeader(int(herr))
					err2 := fmt.Errorf("Could not create homedir permission for %s: %v", userName, err)
					w.Write([]byte(fmt.Sprintf("%v", err2)))
					log.Printf("%v", err2)
					return true
				}
			}
		}
	} else {
		log.Printf("Welcome anonymous user")
	}
	return false
}

func handleFiles(w http.ResponseWriter, r *http.Request, user data.User) bool {
	// If it's a file in our tree...do redirects if we must, or handle a dir reference
	if strings.HasPrefix(r.URL.Path, "/files/") {
		fsPath := path.Dir(r.URL.Path) + "/"
		fsName := path.Base(r.URL.Path)
		attrs := GetAttrs(user, fsPath, fsName)

		// Set headers - we assume that these fields exist
		w.Header().Add("Label", attrs["Label"].(string))
		w.Header().Add("LabelFg", attrs["LabelFg"].(string))
		w.Header().Add("LabelBg", attrs["LabelBg"].(string))
		w.Header().Add("CanRead", ToBoolString(attrs["Read"].(bool)))
		w.Header().Add("CanWrite", ToBoolString(attrs["Write"].(bool)))
		// These are optional
		if v, ok := attrs["ModerationLabel"].(string); ok {
			w.Header().Add("ModerationLabel", v)
		}

		if fs.IsDir(r.URL.Path) {
			isListing := r.URL.Query().Get("listing") == "true"
			fsIndex := r.URL.Path + "index.html"
			if fs.IsExist(fsIndex) && !fs.IsDir(fsIndex) && !isListing {
				fs.ServeFile(w, r, fsIndex)
				return true
			} else {
				dirHandler(w, r)
				return true
			}
		}

		// set a mime type that the handler won't know about if we must
		if strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		}
		if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "text/javascript")
		}
		if strings.HasSuffix(r.URL.Path, ".md") {
			w.Header().Set("Content-Type", "text/markdown")
		}
		if strings.HasSuffix(r.URL.Path, ".templ") {
			w.Header().Set("Content-Type", "text/plain")
		}

		// Recursively find our permissions and return it virtually.
		// It won't show up directly in listings, but should come back
		// with what it finds
		if strings.HasSuffix(r.URL.Path, "--permissions.rego") {
			if fs.IsNotExist(r.URL.Path) {
				r.URL.Path = path.Dir(r.URL.Path) + "/permissions.rego"
			}
		}

		// Walk up the tree until we find what we want
		if path.Base(r.URL.Path) == "permissions.rego" {
			if fs.IsNotExist(r.URL.Path) {
				for true {
					r.URL.Path = path.Dir(path.Dir(r.URL.Path)) + "/permissions.rego"
					if fs.IsExist(r.URL.Path) {
						break // found it!
					}
					if r.URL.Path == "/files/permissions.rego" {
						break
					}
				}
				w.Header().Set("usedfile", r.URL.Path)
			}
		}

		if strings.HasSuffix(r.URL.Path, "--attributes.json") {
			fsPath := path.Dir(r.URL.Path) + "/"
			fsName := path.Base(r.URL.Path)
			w.Write([]byte(utils.AsJson(GetAttrs(user, fsPath, fsName))))
			return true
		}

		// Serve the file we were looking for, possibly with modified URL,
		// set mime types, etc. Special headers could have been set too
		FileServer.ServeHTTP(w, r)
		return true
	}
	return false
}

// Use the standard file serving of Go, because media behavior
// is really really complicated; and you do not want to serve it manually
// if you can help it.
func getHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) {
	// preserve redirect parameters
	q := r.URL.Query().Encode()
	if q != "" {
		q = "?" + q
	}

	if r.URL.Path == "/" {
		getRootHandler(w, r)
		return
	}

	// User hits us with an email link, and we set a cookie
	if r.URL.Path == "/registration/" {
		RegistrationHandler(w, r)
		return
	}

	user := data.GetUser(r)

	if ensureThatHomeDirExists(w, r, user) {
		return
	}

	// User hits us with an email link, and we set a cookie
	if r.URL.Path == "/me" {
		w.Write([]byte(utils.AsJsonPretty(user)))
		return
	}

	// try search handler
	if r.URL.Path == "/search" || strings.HasPrefix(r.URL.Path, "/search/") {
		GetSearchHandler(w, r, pathTokens)
		return
	}

	// Don't deal with directories missing slashes
	if r.URL.Path == "/files" {
		http.Redirect(w, r, r.URL.Path+"/"+q, http.StatusMovedPermanently)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/files/") && fs.IsDir(r.URL.Path) {
		if r.URL.Path[len(r.URL.Path)-1] != '/' {
			http.Redirect(w, r, r.URL.Path+"/"+q, http.StatusMovedPermanently)
			return
		}
	}

	if handleFiles(w, r, user) {
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

// Note that all http handling MUST be tail calls. That makes
// top level routing a little ugly.
func rootRouter(w http.ResponseWriter, r *http.Request) {
	pathTokens := strings.Split(r.URL.Path, "/")
	switch r.Method {
	case http.MethodGet:
		getHandler(w, r, pathTokens)
		return
	case http.MethodPost:
		postHandler(w, r, pathTokens)
		return
	case http.MethodDelete:
		deleteHandler(w, r, pathTokens)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
