package handler

import (
	"fmt"
	"text/template"

	"github.com/rfielding/microcms/fs"
)

func compileTemplate(name string) *template.Template {
	templateText, err := fs.ReadFile(name)
	if err != nil {
		panic(fmt.Sprintf("Cannot find template: %s", name))
	}
	t, err := template.New(name).Parse(string(templateText))
	if err != nil {
		panic(fmt.Sprintf("Cannot parse template %s: %v", name, err))
	}
	return t
}

var compiledRootTemplate *template.Template
var compiledSearchTemplate *template.Template
var compiledListingTemplate *template.Template
var compiledDefaultPermissionsTemplate *template.Template

func TemplatesInit() {
	compiledRootTemplate = compileTemplate("/files/init/rootTemplate.html.templ")
	compiledSearchTemplate = compileTemplate("/files/init/searchTemplate.html.templ")
	compiledListingTemplate = compileTemplate("/files/init/listingTemplate.html.templ")
	compiledDefaultPermissionsTemplate = compileTemplate("/files/init/defaultPermissions.rego.templ")
}
