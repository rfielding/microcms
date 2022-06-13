package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
)

type Node struct {
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Path       string                 `json:"path,omitempty"`
	Name       string                 `json:"name"`
	IsDir      bool                   `json:"isDir"`
	Context    string                 `json:"context,omitempty"`
	Size       int64                  `json:"size,omitempty"`
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
		// TODO: proper relative path calculation
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<form method="GET" action="/search">` + "\n"))
		w.Write([]byte(`<ul>` + "\n"))
		w.Write([]byte(`  <li><label for="match"><input id="match" name="match" type="text"><input type="button" value="search" name="search">` + "\n"))
		w.Write([]byte(`  <li><a href="/files/">files</a>` + "\n"))
		w.Write([]byte(`</ul>` + "\n"))
		w.Write([]byte(`</form>` + "\n"))
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
		attrFileName = fsPath + fName + "--permission.rego"
	} else {
		attrFileName = fsPath + "permission.rego"
	}
	log.Printf("look for permissions at: %s (%s,%s)", attrFileName, fsPath, fName)
	if _, err := os.Stat(attrFileName); err == nil {
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
	if _, err := os.Stat(attrFileName); err == nil {
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

	q := r.URL.Query()
	inJson := q.Get("json") == "true"
	if inJson {
		w.Header().Set("Content-Type", "application/json")
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
		w.Write([]byte(AsJson(listing)))
	} else {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<ul>` + "\n"))
		var prevName string
		for _, name := range names {
			fName := name.Name()

			if strings.HasSuffix(fName, "--thumbnail.png") {
				continue
			}

			// If it's a derived file, then attach it to previous listing
			if strings.HasPrefix(fName, prevName) && strings.Contains(fName, prevName+"--") {
				w.Write([]byte((`  <br>&nbsp;&nbsp;` + "\n")))
			} else {
				w.Write([]byte(`  <br>&nbsp;<li>` + "\n"))
			}

			// Use an image in the link if we have a thumbnail
			sz := getSizeUnits(name.Size(), name.IsDir())

			// Render security attributes
			attrs := getAttrs(user, fsPath, fName)
			if len(attrs) > 0 {
				label, labelOk := attrs["Label"].(string)
				bg, bgOk := attrs["LabelBg"].(string)
				fg, fgOk := attrs["LabelFg"].(string)
				if labelOk && bgOk && fgOk {
					w.Write([]byte(fmt.Sprintf(`<span style="background-color: %s;color: %s">%s</span><br>`+"\n", bg, fg, label)))
				}
			}

			// Render the regular link
			w.Write([]byte(fmt.Sprintf(`<a href="%s">%s %s</a>`+"\n", fName, fName, sz)))

			// Render the thumbnail if we have one
			if _, err := os.Stat(fsPath + "/" + fName + "--thumbnail.png"); err == nil {
				w.Write([]byte(fmt.Sprintf(`<br><a href="%s--thumbnail.png"><img valign=bottom src="%s--thumbnail.png"></a>`+"\n", fName, fName)))
			}

			prevName = fName
		}
		w.Write([]byte(`</ul>` + "\n"))
	}
}
