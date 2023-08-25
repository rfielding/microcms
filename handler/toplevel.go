package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/rfielding/microcms/data"
	"github.com/rfielding/microcms/fs"
	"github.com/rfielding/microcms/utils"
)

// Launch a plain http server
func Setup() {
	bindAddr := utils.Getenv("BIND", "0.0.0.0:9321")
	http.HandleFunc("/", rootRouter)
	log.Printf("start http at %s", bindAddr)
	log.Fatal(http.ListenAndServe(bindAddr, nil))
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

func ToBoolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func writeTimed(w http.ResponseWriter, j []byte) {
	t := MetricsGet.Task()
	t.BytesWrite += int64(len(j))
	defer t.End()
	w.Write(j)
}

func ensureThatHomeDirExists(w http.ResponseWriter, r *http.Request, user data.User) bool {
	// Make home directories exist on first visit
	if user.Identity() != "anonymous" {
		userName := user.Identity()
		parentDir := "/files/" + userName // i can't make it end in slash, seems inconsistent
		fsPath := parentDir + "/"
		fsName := "permissions.rego"
		if !fs.F.IsExist(fsPath + fsName) {
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
		if !strings.HasPrefix(r.URL.Path, "/metrics") {
			log.Printf("Welcome anonymous user")
		}
	}
	return false
}

// Note that all http handling MUST be tail calls. That makes
// top level routing a little ugly.
func rootRouter(w http.ResponseWriter, r *http.Request) {
	pathTokens := strings.Split(r.URL.Path, "/")
	switch r.Method {
	case http.MethodGet:
		task := MetricsGet.Task()
		defer task.End()
		getHandler(w, r, pathTokens)
		return
	case http.MethodPost:
		task := MetricsPost.Task()
		defer task.End()
		postHandler(w, r, pathTokens)
		return
	case http.MethodDelete:
		task := MetricsDelete.Task()
		defer task.End()
		deleteHandler(w, r, pathTokens)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
