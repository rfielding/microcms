package handler

import (
	"net/http"
	"sort"

	"github.com/rfielding/gosqlite/data"
	"github.com/rfielding/gosqlite/fs"
	"github.com/rfielding/gosqlite/utils"
)

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
		w.Write([]byte(utils.AsJson(listing)))
	} else {
		w.Header().Set("Content-Type", "text/html")
		err := compiledRootTemplate.Execute(w, nil)
		if err != nil {
			utils.HandleError(w, err, "Unable to execute rootTemplate: %v", err)
			return
		}
	}
}

func dirHandler(w http.ResponseWriter, r *http.Request) {
	fsPath := r.URL.Path
	user := data.GetUser(r)
	// Get directory names
	names, err := fs.ReadDir(fsPath)
	if err != nil {
		utils.HandleError(w, err, "readdir %s: %v", fsPath)
		return
	}
	sort.Slice(names, func(i, j int) bool {
		return names[i].Name() < names[j].Name()
	})

	listing := data.Listing{
		Children: []data.Node{},
	}
	for _, name := range names {
		fName := name.Name()
		attrs := getAttrs(user, fsPath, fName)
		if attrs["Read"] == true {
			listing.Children = append(listing.Children, data.Node{
				Name:       fName,
				IsDir:      name.IsDir(),
				Size:       name.Size(),
				Attributes: attrs,
			})
		}
	}

	q := r.URL.Query()
	inJson := q.Get("json") == "true"
	if inJson {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(utils.AsJson(listing)))
		return
	} else {
		w.Header().Set("Content-Type", "text/html")
		compiledListingTemplate.Execute(w, listing)
		return
	}
}
