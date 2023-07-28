package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"sort"
	"strings"

	"github.com/rfielding/gosqlite/fs"
)

type Node struct {
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Path       string                 `json:"path,omitempty"`
	Name       string                 `json:"name"`
	IsDir      bool                   `json:"isDir"`
	Context    string                 `json:"context,omitempty"`
	Size       int64                  `json:"size,omitempty"`
	// Used for listings of results
	Part int `json:"part,omitempty"`
}

type Listing struct {
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Children   []Node                 `json:"children"`
}

// Use the same format as the http.FileServer when given a directory
func getRootHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	inJson := q.Get("json") == "true"
	if inJson {
		w.Header().Set("Content-Type", "application/json")
		listing := Listing{
			Children: []Node{
				{Name: "files", IsDir: true},
			},
		}
		w.Write([]byte(AsJson(listing)))
	} else {
		w.Header().Set("Content-Type", "text/html")
		if compiledRootTemplate != nil {
			err := compiledRootTemplate.Execute(w, nil)
			if err != nil {
				HandleError(w, err, "Unable to execute rootTemplate: %v", err)
				return
			}
		} else {
			w.Write([]byte("Please upload /files/init/rootTemplate.html.templ"))
		}
	}
}

// Permission attributes are dynamic, and can come from parent directories.
// The first one found is used to set them all.
// fsPath does NOT begin with a slash, and ends with a slash
func getAttrsPermission(claims interface{}, fsPath string, fName string, initial map[string]interface{}) map[string]interface{} {
	if strings.HasPrefix(fsPath, "./files/") == false {
		panic(fmt.Sprintf("path %s should be rooted under ./files/ and end in slash", fsPath))
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
		jf, err := ioutil.ReadFile(attrFileName)
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
			if fsPath == "./files/" {
				return initial
			} else {
				// careful! if it ends in slash, then parent is same file, fsName is blank!
				parent := "./" + path.Dir(path.Clean(fsPath)) + "/"
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
		jf, err := ioutil.ReadFile(attrFileName)
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

func getSizeUnits(size int64, isDir bool) string {
	sz := ""
	if isDir == false {
		if size > 1024*1024*1024 {
			sz = fmt.Sprintf(" (%d GB)", size/(1024*1024*1024))
		} else if size > 1024*1024 {
			sz = fmt.Sprintf(" (%d MB)", size/(1024*1024))
		} else if size > 1024 {
			sz = fmt.Sprintf(" (%d kB)", size/(1024))
		} else {
			sz = fmt.Sprintf(" (%d B)", size)
		}
	}
	return sz
}

func dirHandler(w http.ResponseWriter, r *http.Request, fsPath string) {
	user := GetUser(r)
	// Get directory names
	names, err := ioutil.ReadDir(fsPath)
	if err != nil {
		HandleError(w, err, "readdir %s: %v", fsPath)
		return
	}
	sort.Slice(names, func(i, j int) bool {
		return names[i].Name() < names[j].Name()
	})

	listing := Listing{
		Children: []Node{},
	}
	for _, name := range names {
		fName := name.Name()
		attrs := getAttrs(user, fsPath, fName)
		listing.Children = append(listing.Children, Node{
			Name:       fName,
			IsDir:      name.IsDir(),
			Size:       name.Size(),
			Attributes: attrs,
		})
	}

	q := r.URL.Query()
	inJson := q.Get("json") == "true"
	if inJson {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(AsJson(listing)))
		return
	} else {
		w.Header().Set("Content-Type", "text/html")
		if compiledListingTemplate != nil {
			compiledListingTemplate.Execute(w, listing)
		} else {
			w.Write([]byte("please upload /files/init/listingTemplate.html.templ"))
		}
		return
	}
}
