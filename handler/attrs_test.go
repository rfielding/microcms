package handler_test

import (
	"testing"

	"github.com/rfielding/microcms/data"
	"github.com/rfielding/microcms/fs"
	"github.com/rfielding/microcms/handler"
)

func TestAttrs(t *testing.T) {
	// Use the persistent mount as the test fixture
	// We assume that the app was brought up and ./deployapps
	fs.At = "../persistent"
	handler.LoadConfig("../config.json")
	handler.TemplatesInit()

	// Test basic attribute calculations
	for k := range data.TheConfig.Users {
		user := data.TheConfig.Users[k]
		email := user["email"][0]
		attrs1 := handler.GetAttrs(user, "/files/init/", "defaultPermissions.rego")
		if attrs1["Read"] != true {
			t.Errorf("Expected read to be true for %s, got %v", email, attrs1["Read"])
		}
		if attrs1["Write"] == true && email == "danica777@gmail.com" {
			t.Errorf("Expected write to be false for %s, got %v", email, attrs1["Write"])
		}
	}
}
