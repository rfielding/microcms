package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/rfielding/microcms/utils"
)

var TheDB *sql.DB

// Setup theDB, and return a cleanup function
func Setup() func() {
	dbName := "persistent/schema.db"
	var err error
	log.Printf("opening database %s", dbName)
	TheDB, err = sql.Open("sqlite3", dbName)
	utils.CheckErr(err, fmt.Sprintf("Could not open %s", dbName))
	log.Printf("opened database %s", dbName)
	return func() {
		TheDB.Close()
		log.Printf("closed database %s", dbName)
	}
}
