package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func IsIn(s string, ss ...string) bool {
	for _, v := range ss {
		if s == v {
			return true
		}
	}
	return false
}

func AsJsonPretty(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("ERR %v", err)
		return ""
	}
	return string(b)
}

func AsJson(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("ERR %v", err)
		return ""
	}
	return string(b)
}

// Use this for startup panics only
func CheckErr(err error, msg string) {
	if err != nil {
		log.Printf("ERR %s", msg)
		panic(err)
	}
}

// Use these on startup so that config is logged
func Getenv(k string, defaultValue string) string {
	v := os.Getenv(k)
	if v == "" {
		v = defaultValue
	}
	log.Printf("ENV %s: %s", k, v)
	return v
}

func GetSizeUnits(size int64, isDir bool) string {
	sz := ""
	if isDir == false {
		if size > 1024*1024*1024 {
			sz = fmt.Sprintf(" (%d GB)", size/(1024*1024*1024))
		} else if size > 1024*1024 {
			sz = fmt.Sprintf(" (%d MB)", size/(1024*1024))
		} else if size > 1024 {
			sz = fmt.Sprintf(" (%d kB)", size/(1024))
		} else {
			sz = fmt.Sprintf(" (%d B)", size)
		}
	}
	return sz
}
