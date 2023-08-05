package main

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/rfielding/microcms/db"
	handler "github.com/rfielding/microcms/handler"
	"github.com/rfielding/microcms/utils"
)

// Make sure to only serve up out of known subdirectories

func main() {
	// In particular, load up the users and config
	handler.LoadConfig()

	handler.DocExtractor = utils.Getenv("DOC_EXTRACTOR", "http://localhost:9998/tika")

	// Set up the database
	dbCleanup := db.Setup()
	defer dbCleanup()

	// this hangs unti the server dies
	handler.Setup()
}