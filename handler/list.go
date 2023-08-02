package handler

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/rfielding/microcms/data"
	"github.com/rfielding/microcms/fs"
	"github.com/rfielding/microcms/utils"
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
		w.Write([]byte(utils.AsJsonPretty(listing)))
	} else {
		w.Header().Set("Content-Type", "text/html")
		err := compiledRootTemplate.Execute(w, nil)
		if err != nil {
			msg := fmt.Sprintf("unable to execute rootTemplate: %v", err)
			log.Printf("ERR %s", msg)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(msg))
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
		msg := fmt.Sprintf("readdir %s: %v", fsPath, err)
		log.Printf("ERR %s", msg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
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
		w.Write([]byte(utils.AsJsonPretty(listing)))
		return
	} else {
		w.Header().Set("Content-Type", "text/html")
		compiledListingTemplate.Execute(w, listing)
		return
	}
}
