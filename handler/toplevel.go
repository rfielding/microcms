package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/rfielding/gosqlite/data"
	"github.com/rfielding/gosqlite/fs"
	"github.com/rfielding/gosqlite/utils"
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
	if len(user["email"]) > 0 {
		// When a user comes in, for rob.fielding@gmail.com
		// /files/rob.fielding@gmail.com/permissions.rego
		// should be uploaded with default permissions if it
		// does not exist. User can override it if permissions
		// must be different.
		// Your first email address is your email.
		userName := user["email"][0]
		parentDir := "/files/" + userName // i can't make it end in slash, seems inconsistent
		fileName := "permissions.rego"
		if !fs.IsExist(parentDir + "/" + fileName) {
			log.Printf("Welcome to %s", parentDir)
			rdr, err := detectNewUser(userName)
			if err != nil {
				utils.HandleReturnedError(w, err, "Could not create homedir for %s: %v", userName)
				return
			}
			if rdr != nil {
				err = postFileHandler(w, r, rdr, parentDir, fileName, parentDir, fileName, false)
				if err != nil {
					utils.HandleReturnedError(w, err, "Could not create homedir permission for %s: %v", userName)
					return
				}
			}
		}
	} else {
		log.Printf("Welcome anonymous user")
	}

	// User hits us with an email link, and we set a cookie
	if r.URL.Path == "/me" {
		w.Write([]byte(utils.AsJson(user)))
		return
	}

	// Don't deal with directories missing slashes
	if r.URL.Path == "/files" {
		http.Redirect(w, r, r.URL.Path+"/"+q, http.StatusMovedPermanently)
		return
	}
	// If it's a file in our tree...do redirects if we must, or handle a dir reference
	if strings.HasPrefix(r.URL.Path, "/files/") {
		if fs.IsDir(r.URL.Path) {
			if r.URL.Path[len(r.URL.Path)-1] != '/' {
				http.Redirect(w, r, r.URL.Path+"/"+q, http.StatusMovedPermanently)
				return
			}
			fsIndex := r.URL.Path + "index.html"
			if fs.IsExist(fsIndex) && !fs.IsDir(fsIndex) {
				fs.ServeFile(w, r, fsIndex)
				return
			} else {
				dirHandler(w, r)
				return
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

		// Serve the file we were looking for, possibly with modified URL,
		// set mime types, etc. Special headers could have been set too
		FileServer.ServeHTTP(w, r)
		return
	}
	// try search handler
	if r.URL.Path == "/search" || strings.HasPrefix(r.URL.Path, "/search/") {
		GetSearchHandler(w, r, pathTokens)
		return
	}
	// give up
	w.WriteHeader(http.StatusNotFound)
}

// We route on method and first segment of the path
func rootRouter(w http.ResponseWriter, r *http.Request) {
	pathTokens := strings.Split(r.URL.Path, "/")
	switch r.Method {
	case http.MethodGet:
		getHandler(w, r, pathTokens)
		return
	case http.MethodPost:
		postHandler(w, r, pathTokens)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
