package main

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/rfielding/gosqlite/fs"
)

func makeThumbnail(file string) (io.Reader, error) {
	command := []string{
		"convert",
		"-thumbnail", "x100",
		"-background", "white",
		"-alpha", "remove",
		"-format", "png",
		(fs.At + file),
		"-",
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

func videoThumbnail(file string) (io.Reader, error) {
	command := []string{
		"convert",
		"-resize", "x100",
		(fs.At + file + "[100]"),
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

func IsVideo(fName string) bool {
	if strings.HasSuffix(fName, ".mp4") {
		return true
	}
	return false
}

func IsImage(fName string) bool {
	if strings.HasSuffix(fName, ".jpg") {
		return true
	}
	if strings.HasSuffix(fName, ".jpeg") {
		return true
	}
	if strings.HasSuffix(fName, ".png") {
		return true
	}
	if strings.HasSuffix(fName, ".gif") {
		return true
	}
	return false
}
