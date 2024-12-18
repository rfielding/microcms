package fs

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os/exec"

	"github.com/rfielding/microcms/utils"
)

/*
 Wrap up os, so that we dont have to make physical paths from logical paths,
 and to remove all the raw stat calls.
*/

type ReadWriteCloser interface {
	io.ReadCloser
	io.WriteCloser
}

// allow for impl substitutions. ie: env vars to use S3, disk otherwise.
var F VFS = NewVolume("./persistent")

type VFS interface {
	Open(fullName string) (io.ReadCloser, error)
	ReadDir(fullName string) ([]fs.DirEntry, error)
	Remove(fullName string) error
	IsExist(fullName string) bool
	IsNotExist(fullName string) bool
	IsDir(fullName string) bool
	Create(fullName string) (io.WriteCloser, error)
	Size(fullName string) int64
	Date(fullName string) string
	ServeFile(w http.ResponseWriter, r *http.Request, fullName string)
	ReadFile(fullName string) ([]byte, error)
	PdfThumbnail(fullName string) (io.Reader, error)
	MakeThumbnail(fullName string) (io.Reader, error)
	VideoThumbnail(fullName string) (io.Reader, error)
	FileServer() http.Handler
}

func commandReader(command []string) (io.Reader, error) {
	cmd := exec.Command(command[0], command[1:]...)
	// This returns an io.ReadCloser, and I don't know if it is mandatory for client to close it
	stdout, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Unable to run command: %v\n%s", err, utils.AsJson(command))
	}
	// Give back a pipe that closes itself when it's read.
	pipeReader, pipeWriter := io.Pipe()
	go func() {
		pipeWriter.Write(stdout)
		pipeWriter.Close()
	}()
	return pipeReader, nil
}
