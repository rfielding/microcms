package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rfielding/microcms/data"
	"github.com/rfielding/microcms/db"
	"github.com/rfielding/microcms/utils"
)

func getSearchHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) {
	lookInside := r.URL.Path[len("/search"):]
	match := r.URL.Query().Get("match")
	rows, err := db.TheDB.Query(`
		SELECT original_path,original_name,part,highlight(filesearch,7,'<b style="background-color:gray">','</b>') highlighted 
		from filesearch
		where filesearch match ? and (original_path || original_name) like ?
	`, match, lookInside+"%")
	if err != nil {
		msg := fmt.Sprintf("query %s: %v", match, err)
		log.Printf("ERR %s", msg)
		w.WriteHeader(http.StatusInternalServerError)
		writeTimed(w, []byte(msg))
		return
	}

	q := r.URL.Query()
	inJson := q.Get("json") == "true"
	hideContent := q.Get("hideContent") == "true"
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
		if hideContent || IsImage(path+name) || IsVideo(path+name) {
			highlighted = ""
		}
		attrs := GetAttrs(user, path, name)
		canRead := attrs.Read
		if canRead {
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
		writeTimed(w, []byte(utils.AsJsonPretty(listing)))
	} else {
		w.Header().Set("Content-Type", "text/html")
		err := compiledSearchTemplate.Execute(w, listing)
		if err != nil {
			msg := fmt.Sprintf("Unable to execute searchTemplate: %v", err)
			log.Printf("ERR %s", msg)
			w.WriteHeader(http.StatusInternalServerError)
			writeTimed(w, []byte(msg))
			return
		}
	}
}
