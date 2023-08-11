package handler

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/rfielding/microcms/data"
	"github.com/rfielding/microcms/fs"
	"github.com/rfielding/microcms/utils"
)

// Use the standard file serving of Go, because media behavior
// is really really complicated; and you do not want to serve it manually
// if you can help it.
func getHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) {
	// preserve redirect parameters
	q := r.URL.Query().Encode()
	if q != "" {
		q = "?" + q
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

	if r.URL.Path == "/" {
		getRootHandler(w, r)
		return
	}

	// User hits us with an email link, and we set a cookie
	if utils.IsIn(r.URL.Path, "/me", "/me/") {
		w.Write([]byte(utils.AsJsonPretty(user)))
		return
	}

	// TODO: GET /search/files/rob.fielding@gmail.com/ to limit to that dir
	if strings.HasPrefix(r.URL.Path, "/search") {
		if r.URL.Path == "/search" {
			r.URL.Path = "/search/"
		}
		getSearchHandler(w, r, pathTokens)
		return
	}

	// Don't deal with directories missing slashes
	if r.URL.Path == "/files" {
		http.Redirect(w, r, r.URL.Path+"/"+q, http.StatusMovedPermanently)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/files/") && fs.F.IsDir(r.URL.Path) {
		if r.URL.Path[len(r.URL.Path)-1] != '/' {
			http.Redirect(w, r, r.URL.Path+"/"+q, http.StatusMovedPermanently)
			return
		}
	}

	if strings.HasPrefix(r.URL.Path, "/metrics") {
		GetMetricsHandler(w, r, pathTokens)
		return
	}

	if handleFiles(w, r, user) {
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func getRootHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	inJson := q.Get("json") == "true"
	if inJson {
		w.Header().Set("Content-Type", "application/json")
		listing := data.Listing{
			Children: []data.Node{
				{Name: "files", IsDir: true},
			},
		}
		writeTimed(w, []byte(utils.AsJsonPretty(listing)))
	} else {
		w.Header().Set("Content-Type", "text/html")
		err := compiledRootTemplate.Execute(w, nil)
		if err != nil {
			msg := fmt.Sprintf("unable to execute rootTemplate: %v", err)
			log.Printf("ERR %s", msg)
			w.WriteHeader(http.StatusInternalServerError)
			writeTimed(w, []byte(msg))
			return
		}
	}
}

func handleFiles(w http.ResponseWriter, r *http.Request, user data.User) bool {
	// If it's a file in our tree...do redirects if we must, or handle a dir reference
	if strings.HasPrefix(r.URL.Path, "/files/") {
		fsPath := path.Dir(r.URL.Path) + "/"
		fsName := path.Base(r.URL.Path)
		attrs := GetAttrs(user, fsPath, fsName)

		// Set headers - we assume that these fields exist
		w.Header().Add("Label", attrs.Label)
		w.Header().Add("LabelFg", attrs.LabelFg)
		w.Header().Add("LabelBg", attrs.LabelBg)
		w.Header().Add("CanRead", ToBoolString(attrs.Read))
		w.Header().Add("CanWrite", ToBoolString(attrs.Write))
		// These are optional
		w.Header().Add("ModerationLabel", attrs.ModerationLabel)

		if fs.F.IsDir(r.URL.Path) {
			isListing := r.URL.Query().Get("listing") == "true"
			fsIndex := r.URL.Path + "index.html"
			if fs.F.IsExist(fsIndex) && !fs.F.IsDir(fsIndex) && !isListing {
				fs.F.ServeFile(w, r, fsIndex)
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

		// TODO: If thumbnail doesn't exist, then return it virtually

		// Recursively find our permissions and return it virtually.
		// It won't show up directly in listings, but should come back
		// with what it finds
		if strings.HasSuffix(r.URL.Path, "--permissions.rego") {
			if fs.F.IsNotExist(r.URL.Path) {
				r.URL.Path = path.Dir(r.URL.Path) + "/permissions.rego"
			}
		}

		// Walk up the tree until we find what we want
		if path.Base(r.URL.Path) == "permissions.rego" {
			if fs.F.IsNotExist(r.URL.Path) {
				for true {
					r.URL.Path = path.Dir(path.Dir(r.URL.Path)) + "/permissions.rego"
					if fs.F.IsExist(r.URL.Path) {
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
			writeTimed(w, []byte(utils.AsJson(GetAttrs(user, fsPath, fsName))))
			return true
		}

		if fs.F.IsExist(r.URL.Path) {
			t := MetricsGet.Task()
			t.BytesWrite += fs.F.Size(r.URL.Path)
			defer t.End()
		}

		// Serve the file we were looking for, possibly with modified URL,
		// set mime types, etc. Special headers could have been set too
		fs.F.FileServer().ServeHTTP(w, r)
		return true
	}
	return false
}
