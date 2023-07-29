package handler

import (
	"log"
	"strings"
)

func IsVideo(name string) bool {
	if strings.HasSuffix(name, ".mp4") {
		return true
	}
	return false
}

func IsImage(name string) bool {
	if strings.HasSuffix(name, ".jpg") {
		return true
	}
	if strings.HasSuffix(name, ".jpeg") {
		return true
	}
	if strings.HasSuffix(name, ".png") {
		return true
	}
	if strings.HasSuffix(name, ".gif") {
		return true
	}
	return false
}

func quote(str string) string {
	if strings.Contains(str, "\"") {
		log.Printf("BEWARE!!! file name has quotes: %s", str)
		return "xxx"
	}
	return "\"" + str + "\""
}

// ie: things that FTS5 can handle directly
func IsTextFile(fName string) bool {
	if strings.HasSuffix(fName, ".txt") {
		return true
	}
	if strings.HasSuffix(fName, ".json") {
		return true
	}
	if strings.HasSuffix(fName, ".html") {
		return true
	}
	if strings.HasSuffix(fName, ".vtt") {
		return true
	}
	return false
}

// ie: things that Tika can handle to produce IsTextFile
func IsDoc(fName string) bool {
	if strings.HasSuffix(fName, ".doc") {
		return true
	}
	if strings.HasSuffix(fName, ".ppt") {
		return true
	}
	if strings.HasSuffix(fName, ".xls") {
		return true
	}
	if strings.HasSuffix(fName, ".docx") {
		return true
	}
	if strings.HasSuffix(fName, ".pptx") {
		return true
	}
	if strings.HasSuffix(fName, ".xlsx") {
		return true
	}
	if strings.HasSuffix(fName, ".pdf") {
		return true
	}
	// ?? a guess
	if strings.HasSuffix(fName, ".one") {
		return true
	}
	return false
}
