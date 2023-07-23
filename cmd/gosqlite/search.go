package main

import (
	"net/http"
)

func getSearchHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) {
	match := r.URL.Query().Get("match")
	rows, err := theDB.Query(`
		SELECT original_path,original_name,part,highlight(filesearch,7,'<b style="background-color:yellow">','</b>') highlighted 
		from filesearch
		where filesearch match ?
	`, match)
	if err != nil {
		HandleError(w, err, "query %s: %v", match)
		return
	}

	q := r.URL.Query()
	inJson := q.Get("json") == "true"
	listing := Listing{
		Children: []Node{
			{Name: "files", IsDir: true},
		},
	}
	for rows.Next() {
		var path, name, highlighted string
		var part int
		rows.Scan(&path, &name, &part, &highlighted)
		listing.Children = append(listing.Children, Node{
			Path:    path,
			Name:    name,
			Part:    part,
			IsDir:   false,
			Context: highlighted,
		})
	}

	if inJson {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(AsJson(listing)))
	} else {
		w.Header().Set("Content-Type", "text/html")
		if compiledSearchTemplate != nil {
			err := compiledSearchTemplate.Execute(w, listing)
			if err != nil {
				HandleError(w, err, "Unable to execute searchTemplate: %v", err)
				return
			}
		} else {
			w.Write([]byte("Please upload /files/searchTemplate.html.templ"))
		}
		/*
			w.Write([]byte(`<ul>` + "\n"))
			for rows.Next() {
				var path, name, highlighted string
				var part int
				rows.Scan(&path, &name, &part, &highlighted)
				w.Write([]byte(
					fmt.Sprintf(`<li><a href="%s%s">%s%s [part %d]</a><br>%s`+"<br></li>", path, name, path, name, part, highlighted),
				))
			}
			w.Write([]byte(`</ul>`)
		*/
	}
}
