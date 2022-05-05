package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

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

// XXX
// will take a filestream and use better heuristics later.
// the point is to quickly get something indexed upon upload.
func isTextFile(name string) bool {
	if strings.HasSuffix(name, ".txt") {
		return true
	}
	if strings.HasSuffix(name, ".json") {
		return true
	}
	if strings.HasSuffix(name, ".html") {
		return true
	}
	return false
}

func indexTextFile(command string, path string, name string, part int, content []byte) error {
	// index the file
	_, err := theDB.Exec(
		`INSERT INTO filesearch (cmd, path, name, part, content) VALUES (?, ?, ?, ?)`,
		command,
		path,
		name,
		part,
		content,
	)
	if err != nil {
		return fmt.Errorf("ERR while indexing %s %s%s: %v", command, path, name, err)
	}
	return nil
}

func postFilesHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) {
	var err error
	defer r.Body.Close()

	if len(pathTokens) < 2 {
		msg := fmt.Sprintf("path needs /[command]/[url] for post to %s: %v", r.URL.Path, err)
		log.Printf("ERR %s", msg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return
	}
	command := pathTokens[1]
	pathTokens[1] = "files"
	// Write the file out
	flags := os.O_WRONLY | os.O_CREATE

	// We either append content, or overwrite it entirely
	if command == "append" {
		flags |= os.O_APPEND
	}
	if command == "files" {
		flags |= os.O_TRUNC
	}

	// TODO: check permissions before allowing writes

	// Make sure that the path exists, and get the file name
	parentDir := strings.Join(pathTokens[:len(pathTokens)-1], "/")
	name := pathTokens[len(pathTokens)-1]

	log.Printf("Ensure existence of parentDir: %s", parentDir)
	err = os.MkdirAll("."+parentDir, 0777)
	if err != nil {
		msg := fmt.Sprintf("Could not create path for %s: %v", r.URL.Path, err)
		log.Printf("ERR %s", msg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return
	}

	f, err := os.OpenFile("."+r.URL.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		msg := fmt.Sprintf("Could not create file %s: %v", r.URL.Path, err)
		log.Printf("ERR %s", msg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return
	}

	sz, err := io.Copy(f, r.Body)
	f.Close() // strange positioning, but we must close before defer can get to it.
	if err != nil {
		msg := fmt.Sprintf("Could not write to file (%d bytes written) %s: %v", sz, r.URL.Path, err)
		log.Printf("ERR %s", msg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return
	}

	if isTextFile(r.URL.Path) {
		// open the file that we saved, and index it in the database.
		f, err := os.Open("." + r.URL.Path)
		if err != nil {
			msg := fmt.Sprintf("Could not open file for indexing %s: %v", r.URL.Path, err)
			log.Printf("ERR %s", msg)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(msg))
			return
		}
		defer f.Close()
		buffer := make([]byte, 1024*1024)
		part := 0
		for {
			sz, err := f.Read(buffer)
			if err == io.EOF {
				break
			}
			err = indexTextFile(command, parentDir+"/", name, part, buffer[:sz])
			if err != nil {
				log.Printf("failed indexing: %v", err)
			}
			part++
		}
	}
}

func postHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) {
	if len(pathTokens) > 2 && pathTokens[1] == "files" {
		postFilesHandler(w, r, pathTokens)
		return
	}
	w.WriteHeader(http.StatusNotImplemented)
}

// Use the same format as the http.FileServer when given a directory
func getRootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<pre>` + "\n"))
	w.Write([]byte(`<a href="files">files</a>` + "\n"))
	w.Write([]byte(`</pre>`))
}

// Use the standard file serving of Go, because media behavior
// is really really complicated; and you do not want to serve it manually
// if you can help it.
func getHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) {
	if len(pathTokens) > 1 && pathTokens[1] == "" {
		getRootHandler(w, r)
		return
	}
	if len(pathTokens) > 1 && pathTokens[1] == "files" {
		theFS.ServeHTTP(w, r)
		return
	}
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
	// Set up the database
	dbCleanup := dbSetup()
	defer dbCleanup()

	// this hangs unti the server dies
	httpSetup()
}
