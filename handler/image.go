package handler

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/rfielding/microcms/fs"
	"github.com/rfielding/microcms/utils"
)

func MakeThumbnail(file string) (io.Reader, error) {
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
		return nil, fmt.Errorf("Unable to run thumbnail command: %v\n%s", err, utils.AsJson(command))
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
		return nil, fmt.Errorf("Unable to run thumbnail command: %v\n%s", err, utils.AsJson(command))
	}
	// Give back a pipe that closes itself when it's read.
	pipeReader, pipeWriter := io.Pipe()
	go func() {
		pipeWriter.Write(stdout)
		pipeWriter.Close()
	}()
	return pipeReader, nil
}
