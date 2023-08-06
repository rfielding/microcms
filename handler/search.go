package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/rfielding/microcms/data"
	"github.com/rfielding/microcms/db"
	"github.com/rfielding/microcms/fs"
	"github.com/rfielding/microcms/utils"
)

// Permission attributes are dynamic, and can come from parent directories.
// The first one found is used to set them all.
// fsPath does NOT begin with a slash, and ends with a slash
func getAttrsPermission(claims data.User, fsPath string, fsName string, initial map[string]interface{}) map[string]interface{} {
	// Try exact file if fName is not blank
	regoFile := fsPath + "permissions.rego"
	if fsName != "" {
		regoFile = fsPath + fsName + "--permissions.rego"
		if fs.IsDir(fsPath + fsName) {
			regoFile = fsPath + fsName + "/permissions.rego"
		}
	}
	if fs.IsExist(regoFile) {
		jf, err := fs.ReadFile(regoFile)
		if err != nil {
			log.Printf("Failed to open %s!: %v", regoFile, err)
		} else {
			regoString := string(jf)
			calculation, err := CalculateRego(claims, regoString)
			if err != nil {
				log.Printf("Failed to parse rego %s!: %v\n%s", regoFile, err, regoString)
			}
			for k, v := range calculation {
				initial[k] = v
			}
		}
		return initial
	} else {
		if fsName != "" {
			return getAttrsPermission(claims, fsPath, "", initial)
		} else {
			if fsPath == "/files/" {
				return initial
			} else {
				// careful! if it ends in slash, then parent is same file, fsName is blank!
				fsPath := path.Dir(path.Clean(fsPath)) + "/"
				fsName := ""
				return getAttrsPermission(claims, fsPath, fsName, initial)
			}
		}
	}
}

func GetAttrs(claims data.User, fsPath string, fsName string) map[string]interface{} {
	// always get attributes according to the original file
	if strings.Contains(fsName, "--") {
		//a better pattern would be:  *.*--*.*
		fNameOriginal := fsName[0:strings.LastIndex(fsName, "--")]
		fsName = fNameOriginal
	}
	attrs := make(map[string]interface{})

	// Start parsing attributes with a custom set of values that
	// get overridden with calculated values
	customFileName := fsPath + fsName + "--custom.json"
	if fs.IsExist(customFileName) {
		jf, err := fs.ReadFile(customFileName)
		if err != nil {
			log.Printf("Failed to open %s!: %v", customFileName, err)
		} else {
			err := json.Unmarshal(jf, &attrs)
			if err != nil {
				log.Printf("Failed to parse json %s!: %v", customFileName, err)
			}
		}
	}
	// If there is content moderation, then add it in here
	mods := make(map[string]interface{})
	moderationFileName := fsPath + fsName + "--moderation.json"
	if fs.IsExist(moderationFileName) {
		jf, err := fs.ReadFile(moderationFileName)
		if err != nil {
			log.Printf("Failed to open %s!: %v", customFileName, err)
		} else {
			err := json.Unmarshal(jf, &mods)
			if err != nil {
				log.Printf("Failed to parse json %s!: %v", customFileName, err)
			}
			modsList, ok := mods["ModerationLabels"].([]interface{})
			if ok && len(modsList) > 0 {
				attrs["Moderation"] = true
				modsObj, ok := modsList[0].(map[string]interface{})
				if ok {
					val, ok := modsObj["Name"].(string)
					if ok {
						attrs["ModerationLabel"] = val
					}
				}
			}
		}
	}
	// overwrite with calculated values
	return getAttrsPermission(claims, fsPath, fsName, attrs)
}

func GetSearchHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) {
	lookInside := r.URL.Path[len("/search"):]
	match := r.URL.Query().Get("match")
	rows, err := db.TheDB.Query(`
		SELECT original_path,original_name,part,highlight(filesearch,7,'<b style="background-color:gray">','</b>') highlighted 
		from filesearch
		where filesearch match ? and original_path like ?
	`, match, lookInside+"%")
	if err != nil {
		msg := fmt.Sprintf("query %s: %v", match, err)
		log.Printf("ERR %s", msg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return
	}

	q := r.URL.Query()
	inJson := q.Get("json") == "true"
	listing := data.Listing{
		Node: data.Node{
			Path:       "/",
			Name:       "files",
			IsDir:      true,
			Attributes: GetAttrs(data.GetUser(r), "/", "files"),
		},
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
		attrs := GetAttrs(user, path, name)
		canRead, ok := attrs["Read"].(bool)
		if ok && canRead {
			listing.Children = append(listing.Children, data.Node{
				Path:       path,
				Name:       name,
				Part:       part,
				IsDir:      false,
				Context:    highlighted,
				Attributes: attrs,
			})
		}
	}

	if inJson {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(utils.AsJsonPretty(listing)))
	} else {
		w.Header().Set("Content-Type", "text/html")
		err := compiledSearchTemplate.Execute(w, listing)
		if err != nil {
			msg := fmt.Sprintf("Unable to execute searchTemplate: %v", err)
			log.Printf("ERR %s", msg)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(msg))
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
