package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var docExtractor string

// Make sure to only serve up out of known subdirectories
var theFS = http.FileServer(http.Dir("."))
var theDB *sql.DB

// Use this for startup panics only
func CheckErr(err error, msg string) {
	if err != nil {
		log.Printf("ERR %s", msg)
		panic(err)
	}
}

// Use these on startup so that config is logged
func Getenv(k string, defaultValue string) string {
	v := os.Getenv(k)
	if v == "" {
		v = defaultValue
	}
	log.Printf("ENV %s: %s", k, v)
	return v
}

// ie: things that FTS5 can handle directly
func IsTextFile(fName string) bool {
	if strings.HasSuffix(fName, ".txt") {
		return true
	}
	if strings.HasSuffix(fName, ".json") {
		return true
	}
	if strings.HasSuffix(fName, ".html") {
		return true
	}
	if strings.HasSuffix(fName, ".vtt") {
		return true
	}
	return false
}

func AsJson(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("ERR %v", err)
		return ""
	}
	return string(b)
}

func postHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) {
	if len(pathTokens) > 2 && pathTokens[1] == "files" {
		postFilesHandler(w, r, pathTokens)
		return
	}
	w.WriteHeader(http.StatusNotImplemented)
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

	user := GetUser(r)
	if len(user["email"]) > 0 {
		log.Printf("Welcome user: %s", AsJson(user))
	} else {
		log.Printf("Welcome anonymous user")
	}

	// Don't deal with directories missing slashes
	if r.URL.Path == "/files" {
		http.Redirect(w, r, r.URL.Path+"/"+q, http.StatusMovedPermanently)
		return
	}
	// If it's a file in our tree...do redirects if we must, or handle a dir reference
	if strings.HasPrefix(r.URL.Path, "/files/") {
		s, _ := os.Stat("." + r.URL.Path)
		if s != nil && s.IsDir() {
			if r.URL.Path[len(r.URL.Path)-1] != '/' {
				http.Redirect(w, r, r.URL.Path+"/"+q, http.StatusMovedPermanently)
				return
			}
			sIdx, _ := os.Stat("." + r.URL.Path + "index.html")
			if sIdx != nil && !sIdx.IsDir() {
				// Rather than redirect?
				http.ServeFile(w, r, "."+r.URL.Path+"index.html")
				return
			} else {
				dirHandler(w, r, "."+r.URL.Path)
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
			if _, err := os.Stat("." + r.URL.Path); os.IsNotExist(err) {
				r.URL.Path = path.Dir(r.URL.Path) + "/permissions.rego"
			}
		}

		// Walk up the tree until we find what we want
		if path.Base(r.URL.Path) == "permissions.rego" {
			if _, err := os.Stat("." + r.URL.Path); os.IsNotExist(err) {
				for true {
					r.URL.Path = path.Dir(path.Dir(r.URL.Path)) + "/permissions.rego"
					_, err = os.Stat("." + r.URL.Path)
					if os.IsExist(err) {
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
		theFS.ServeHTTP(w, r)
		return
	}
	// try search handler
	if r.URL.Path == "/search" || strings.HasPrefix(r.URL.Path, "/search/") {
		getSearchHandler(w, r, pathTokens)
		return
	}
	// give up
	w.WriteHeader(http.StatusNotFound)
}

func HandleError(w http.ResponseWriter, err error, mask string, args ...interface{}) {
	msg := fmt.Sprintf(mask, append(args, err.Error())...)
	log.Printf("ERR %s", msg)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(msg))
}

func HandleReturnedError(w http.ResponseWriter, err error, mask string, args ...interface{}) error {
	msg := fmt.Sprintf(mask, append(args, err.Error())...)
	log.Printf("ERR %s", msg)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(msg))
	return fmt.Errorf("%v", msg)
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

// Setup theDB, and return a cleanup function
func dbSetup() func() {
	var err error
	dbName := Getenv("SCHEMA", "schema.db")
	log.Printf("opening database %s", dbName)
	theDB, err = sql.Open("sqlite3", dbName)
	CheckErr(err, fmt.Sprintf("Could not open %s", dbName))
	log.Printf("opened database %s", dbName)
	return func() {
		theDB.Close()
		log.Printf("closed database %s", dbName)
	}
}

// Launch a plain http server
func httpSetup() {
	bindAddr := Getenv("BIND", "0.0.0.0:9321")
	http.HandleFunc("/", rootRouter)
	log.Printf("start http at %s", bindAddr)
	log.Fatal(http.ListenAndServe(bindAddr, nil))
}

func main() {
	// In particular, load up the users and config
	LoadConfig()

	docExtractor = Getenv("DOC_EXTRACTOR", "http://localhost:9998/tika")

	// Set up the database
	dbCleanup := dbSetup()
	defer dbCleanup()

	// this hangs unti the server dies
	httpSetup()
}
