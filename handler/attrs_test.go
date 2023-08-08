package handler_test

import (
	"log"
	"os"
	"testing"

	"github.com/rfielding/microcms/data"
	"github.com/rfielding/microcms/fs"
	"github.com/rfielding/microcms/handler"
	"github.com/rfielding/microcms/utils"
)

func TestMetrics(t *testing.T) {
	c := handler.NewMetricsCollector("test")
	m := c.Task()
	m.BytesRead += 20
	m.BytesWrite += 33
	m.End()
	s := c.Stats()
	t.Logf("Stats: %s", utils.AsJson(s))

	handler.GetMetricsWriter(os.Stdout)
}

func TestAttrs(t *testing.T) {
	// Use the persistent mount as the test fixture
	// We assume that the app was brought up and ./deployapps
	fs.At = "../persistent"
	handler.LoadConfig("../config.json")
	handler.TemplatesInit()

	// Test basic attribute calculations
	var userRob data.User
	var userDanica data.User
	for k := range data.TheConfig.Users {
		user := data.TheConfig.Users[k]
		email := user["email"][0]
		attrs1 := handler.GetAttrs(user, "/files/init/", "defaultPermissions.rego")
		if attrs1.Read != true {
			t.Errorf("Expected read to be true for %s, got %v", email, attrs1.Read)
		}
		if attrs1.Write == true && email == "danica777@gmail.com" {
			t.Errorf("Expected write to be false for %s, got %v", email, attrs1.Write)
		}
		if email == "rob.fielding@gmail.com" {
			userRob = user
		}
		if email == "danica777@gmail.com" {
			userDanica = user
		}
	}
	_ = userDanica
	attrsCat := handler.GetAttrs(
		userRob,
		"/files/rob.fielding@gmail.com/documents/",
		"ktt.jpg",
	)
	log.Printf("attrsCat: %s", utils.AsJson(attrsCat))

	a1 := handler.GetAttrs(
		userRob,
		"/files/rob.fielding@gmail.com/documents/",
		"nm.jpg",
	)
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" {
		log.Printf("a1: %s", utils.AsJson(a1))
		a2 := handler.GetAttrs(
			userRob,
			"/files/rob.fielding@gmail.com/documents/",
			"nm.jpg--thumbnail.jpg",
		)
		log.Printf("a2: %s", utils.AsJson(a2))
	}

}
