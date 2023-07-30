package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/rfielding/gosqlite/data"
	"github.com/rfielding/gosqlite/db"
	"github.com/rfielding/gosqlite/fs"
	"github.com/rfielding/gosqlite/utils"
)

// Permission attributes are dynamic, and can come from parent directories.
// The first one found is used to set them all.
// fsPath does NOT begin with a slash, and ends with a slash
func getAttrsPermission(claims interface{}, fsPath string, fName string, initial map[string]interface{}) map[string]interface{} {
	if strings.HasPrefix(fsPath, "/files/") == false {
		panic(fmt.Sprintf("path %s should be rooted under /files/ and end in slash", fsPath))
	}
	// Try exact file if fName is not blank
	attrFileName := ""
	if fName != "" {
		attrFileName = fsPath + fName + "--permissions.rego"
	} else {
		attrFileName = fsPath + "permissions.rego"
	}
	//log.Printf("look for permissions at: %s (%s,%s)", attrFileName, fsPath, fName)
	if fs.IsExist(attrFileName) {
		jf, err := fs.ReadFile(attrFileName)
		if err != nil {
			log.Printf("Failed to open %s!: %v", attrFileName, err)
		} else {
			calculation, err := CalculateRego(claims, string(jf))
			if err != nil {
				log.Printf("Failed to parse %s!: %v", attrFileName, err)
			}
			for k, v := range calculation {
				initial[k] = v
			}
		}
		return initial
	} else {
		if fName != "" {
			return getAttrsPermission(claims, fsPath, "", initial)
		} else {
			if fsPath == "/files/" {
				return initial
			} else {
				// careful! if it ends in slash, then parent is same file, fsName is blank!
				parent := path.Dir(path.Clean(fsPath)) + "/"
				return getAttrsPermission(claims, parent, "", initial)
			}
		}
	}
}

func getAttrs(claims interface{}, fsPath string, fName string) map[string]interface{} {
	// Get the attributes for the file if they exist
	attrs := make(map[string]interface{})
	attrFileName := fsPath + fName + "--attributes.json"
	if fs.IsExist(attrFileName) {
		jf, err := fs.ReadFile(attrFileName)
		if err != nil {
			log.Printf("Failed to open %s!: %v", attrFileName, err)
		} else {
			err := json.Unmarshal(jf, &attrs)
			if err != nil {
				log.Printf("Failed to parse %s!: %v", attrFileName, err)
			}
		}
	}
	return getAttrsPermission(claims, fsPath, fName, attrs)
}

func GetSearchHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) {
	match := r.URL.Query().Get("match")
	rows, err := db.TheDB.Query(`
		SELECT original_path,original_name,part,highlight(filesearch,7,'<b style="background-color:gray">','</b>') highlighted 
		from filesearch
		where filesearch match ?
	`, match)
	if err != nil {
		utils.HandleError(w, err, "query %s: %v", match)
		return
	}

	q := r.URL.Query()
	inJson := q.Get("json") == "true"
	listing := data.Listing{
		Children: []data.Node{},
	}
	user := data.GetUser(r)
	for rows.Next() {
		var path, name, highlighted string
		var part int
		rows.Scan(&path, &name, &part, &highlighted)
		if IsImage(path+name) || IsVideo(path+name) {
			highlighted = ""
		}
		listing.Children = append(listing.Children, data.Node{
			Path:       path,
			Name:       name,
			Part:       part,
			IsDir:      false,
			Context:    highlighted,
			Attributes: getAttrs(user, path, name),
		})
	}

	if inJson {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(utils.AsJson(listing)))
	} else {
		w.Header().Set("Content-Type", "text/html")
		err := compiledSearchTemplate.Execute(w, listing)
		if err != nil {
			utils.HandleError(w, err, "Unable to execute searchTemplate: %v", err)
			return
		}
	}
}

func indexTextFile(
	path string,
	name string,
	part int,
	originalPath string,
	originalName string,
	content []byte,
) error {
	// index the file -- if we are appending, we should only incrementally index
	_, err := db.TheDB.Exec(
		`INSERT INTO filesearch (cmd, path, name, part, original_path, original_name, content) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"files",
		path,
		name,
		part,
		originalPath,
		originalName,
		content,
	)
	if err != nil {
		return fmt.Errorf("ERR while indexing files %s%s: %v", path, name, err)
	}
	return nil
}
