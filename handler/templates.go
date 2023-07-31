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

// Defaults were uploaded into the container
var compiledRootTemplate *template.Template = compileTemplate("/files/init/rootTemplate.html.templ")
var compiledSearchTemplate *template.Template = compileTemplate("/files/init/searchTemplate.html.templ")
var compiledListingTemplate *template.Template = compileTemplate("/files/init/listingTemplate.html.templ")
var compiledDefaultPermissionsTemplate *template.Template = compileTemplate("/files/init/defaultPermissions.rego.templ")
