package main

import (
	"archive/tar"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
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
	// index the file -- if we are appending, we should only incrementally index
	_, err := theDB.Exec(
		`INSERT INTO filesearch (cmd, path, name, part, content) VALUES (?, ?, ?, ?, ?)`,
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

// postFileHandler can be re-used as long as err != nil
func postFileHandler(
	w http.ResponseWriter,
	r *http.Request,
	stream io.Reader,
	command string,
	parentDir string,
	name string,
) error {
	fullName := fmt.Sprintf("%s/%s", parentDir, name)
	log.Printf("create %s %s", command, fullName)

	// Write the file out
	flags := os.O_WRONLY | os.O_CREATE

	// We either append content, or overwrite it entirely
	if command == "append" {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	// TODO: check permissions before allowing writes

	log.Printf("Ensure existence of parentDir: %s", parentDir)
	err := os.MkdirAll("."+parentDir, 0777)
	if err != nil {
		msg := fmt.Sprintf("Could not create path for %s: %v", r.URL.Path, err)
		log.Printf("ERR %s", msg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return fmt.Errorf("%v", msg)
	}

	existingSize := int64(0)
	s, err := os.Stat("." + fullName)
	if err == nil {
		existingSize = s.Size()
	}

	// XXX forget about append for a moment... just get it right on regular uploads

	f, err := os.Create("." + fullName) //, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		msg := fmt.Sprintf("Could not create file %s: %v", r.URL.Path, err)
		log.Printf("ERR %s", msg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return fmt.Errorf("%v", msg)
	}

	// Save the stream to a file
	sz, err := io.Copy(f, stream)
	f.Close() // strange positioning, but we must close before defer can get to it.
	if err != nil {
		msg := fmt.Sprintf("Could not write to file (%d bytes written) %s: %v", sz, r.URL.Path, err)
		log.Printf("ERR %s", msg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return fmt.Errorf("%v", msg)
	}

	if isTextFile(fullName) {
		// open the file that we saved, and index it in the database.
		f, err := os.Open("." + fullName)
		if err != nil {
			msg := fmt.Sprintf("Could not open file for indexing %s: %v", fullName, err)
			log.Printf("ERR %s", msg)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(msg))
			return fmt.Errorf("%v", msg)
		}
		defer f.Close()
		if existingSize > 0 {
			// we are appending, so we need to start at the end of the file
			f.Seek(existingSize, 0)
		}
		if command == "files" {
			// this implies a truncate
			_, err := theDB.Exec(`DELETE from filesearch where path = ? and name = ? and cmd = ?`, parentDir+"/", name, command)
			if err != nil {
				log.Printf("cleaning out fulltextsearch for: %s%s %s failed: %v", parentDir+"/", name, command, err)
			}
		}
		buffer := make([]byte, 4*1024)
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
	return nil
}

func postFilesHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) {
	var err error
	defer r.Body.Close()

	q := r.URL.Query()
	// This is a signal that this is a tar archive
	// that we unpack to install all files at the given url
	needsInstall := q.Get("install") == "true"
	if needsInstall {
		log.Printf("install tarball to %s", r.URL.Path)
	}

	if len(pathTokens) < 2 {
		msg := fmt.Sprintf("path needs /[command]/[url] for post to %s: %v", r.URL.Path, err)
		log.Printf("ERR %s", msg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return
	}
	command := pathTokens[1]
	pathTokens[1] = "files"

	// Make sure that the path exists, and get the file name
	parentDir := strings.Join(pathTokens[:len(pathTokens)-1], "/")
	name := pathTokens[len(pathTokens)-1]

	// If err != nil, then we can't call this again.  http response has been sent
	if needsInstall == true { // we ASSUME that it's a tar
		// We install the tarball, walking it file by file
		t := tar.NewReader(r.Body)
		for {
			header, err := t.Next()
			if err == io.EOF {
				break
			}
			// Ignore directories for a moment XXX
			if header.Typeflag == tar.TypeReg {
				// I assume that header names are unqualified dir names
				tname := strings.Split(header.Name, "/") // expect . in front
				tname = tname[1:]
				tardir := path.Dir(fmt.Sprintf("%s/%s/%s", parentDir, name, strings.Join(tname, "/")))
				tarname := path.Base(header.Name)
				log.Printf("writing: %s into %s", tarname, tardir)
				err = postFileHandler(w, r, t, command, tardir, tarname)
				if err != nil {
					log.Printf("ERR %v", err)
					return
				}
			}
		}
	} else {
		// Just a normal single-file upload
		err = postFileHandler(w, r, r.Body, command, parentDir, name)
		if err != nil {
			log.Printf("ERR %v", err)
			return
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

func getSearchHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) {
	match := r.URL.Query().Get("match")
	rows, err := theDB.Query(`
		SELECT path,name,part,highlight(filesearch,5,'<b style="background-color:yellow">','</b>') highlighted 
		from filesearch
		where filesearch match ?
	`, match)
	if err != nil {
		msg := fmt.Sprintf("query %s: %v", match, err)
		log.Printf("ERR %s", msg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<ul>` + "\n"))
	for rows.Next() {
		var path, name, highlighted string
		var part int
		rows.Scan(&path, &name, &part, &highlighted)
		w.Write([]byte(
			fmt.Sprintf(`<li><a href="%s%s">%s%s [part %d]</a><br>%s`+"<br></li>", path, name, path, name, part, highlighted),
		))
	}
	w.Write([]byte(`</ul>`))
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
		if strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		}
		if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "text/javascript")
		}
		if strings.HasSuffix(r.URL.Path, ".md") {
			w.Header().Set("Content-Type", "text/markdown")
		}
		theFS.ServeHTTP(w, r)
		return
	}
	if len(pathTokens) > 1 && pathTokens[1] == "search" {
		getSearchHandler(w, r, pathTokens)
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
