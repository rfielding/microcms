package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/rfielding/gosqlite/fs"
)

func quote(str string) string {
	if strings.Contains(str, "\"") {
		log.Printf("BEWARE!!! file name has quotes: %s", str)
		return "xxx"
	}
	return "\"" + str + "\""
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

func pdfThumbnail(file string) (io.Reader, error) {
	command := []string{
		"convert",
		"-resize", "x100",
		(fs.At + file + "[0]"),
		"png:-",
	}
	cmd := exec.Command(command[0], command[1:]...)
	// This returns an io.ReadCloser, and I don't know if it is mandatory for client to close it
	stdout, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Unable to run thumbnail command: %v\n%s", err, AsJson(command))
	}
	// Give back a pipe that closes itself when it's read.
	pipeReader, pipeWriter := io.Pipe()
	go func() {
		pipeWriter.Write(stdout)
		pipeWriter.Close()
	}()
	return pipeReader, nil
}

// Make a request to tika in this case
func DocExtract(fName string, rdr io.Reader) (io.ReadCloser, error) {
	cl := http.Client{}
	req, err := http.NewRequest("PUT", docExtractor, rdr)
	if err != nil {
		return nil, fmt.Errorf("Unable to make request to upload file: %v", err)
	}
	req.Header.Add("accept", "text/plain")
	res, err := cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Unable to do request to upload file %s: %v", fName, err)
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Unable to upload %s: %d", fName, res.StatusCode)
	}
	return res.Body, nil
}
