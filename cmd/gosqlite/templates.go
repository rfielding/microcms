package main

import (
	"fmt"
	"io/ioutil"
	"text/template"
)

func compileTemplate(name string) *template.Template {
	templateText, err := ioutil.ReadFile(name)
	if err != nil {
		panic(fmt.Sprintf("Cannot find template: %s", name))
	}
	t, err := template.New(name).Parse(string(templateText))
	if err != nil {
		panic(fmt.Sprintf("Cannot parse template %s: %v", name, err))
	}
	return t
}

// You should upload these templates as part of initialization of the CMS
var compiledRootTemplate *template.Template
var compiledSearchTemplate *template.Template
var compiledListingTemplate *template.Template
