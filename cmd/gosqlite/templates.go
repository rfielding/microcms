package main

import (
	"fmt"
	"text/template"
)

var searchTemplate = `
<form method="GET" action="/search">
  <ul>
    <li><label for="match">Search:<input id="match" name="match" type="text">
    <li><a href="/files/">files</a>
  </ul>
</form>
`

func compileSearchTemplate() *template.Template {
	t, err := template.New("searchTemplate").Parse(searchTemplate)
	if err != nil {
		panic(fmt.Sprintf("Cannot parse template: %v", err))
	}
	return t
}

var compiledSearchTemplate = compileSearchTemplate()
