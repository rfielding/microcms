package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rfielding/microcms/data"
	"github.com/rfielding/microcms/db"
	"github.com/rfielding/microcms/handler"
	"github.com/rfielding/microcms/utils"
)

func loadUsers() {
	cfgfile := "./config.json"
	lastHash := []byte("")
	for {
		// load up the file as raw bytes
		cfg, err := os.ReadFile(cfgfile)
		if err != nil {
			log.Printf("cannot read %s: %v", cfgfile, err)
		}

		// try to just parse it as json for a config
		testCfg := data.Config{}
		err = json.Unmarshal(cfg, &testCfg)
		if err != nil {
			log.Printf("cannot parse %s: %v", cfgfile, err)
		} else {
			h := sha256.Sum256(cfg)
			newHash := h[:]
			if !bytes.Equal(newHash, lastHash) {
				handler.LoadConfig(cfgfile)
			}
			lastHash = newHash
		}

		time.Sleep(time.Duration(5) * time.Second)
	}
}

func main() {
	go loadUsers()

	handler.TemplatesInit()

	handler.DocExtractor = utils.Getenv("DOC_EXTRACTOR", "http://localhost:9998/tika")

	// Set up the database
	dbCleanup := db.Setup()
	defer dbCleanup()

	// this hangs until the server dies
	handler.Setup()
}
