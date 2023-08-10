package fs

import (
	"io"
	"io/fs"
	"net/http"
)

/*
 Wrap up os, so that we dont have to make physical paths from logical paths,
 and to remove all the raw stat calls.
*/

type ReadWriteCloser interface {
	io.ReadCloser
	io.WriteCloser
}

// Handling At() implies that the local conversion programs may move into here,
// because they use `name`
// This probably means a way of alloc/free temp files for the duraction of
// operations that must operate with the actual filesystem.

type VFS interface {
	At() string // hmm. still have filesystem dependencies due to ImageMagick, etc.
	Open(fullName string) (io.ReadCloser, error)
	ReadDir(fullName string) ([]fs.DirEntry, error)
	Remove(fullName string) error
	IsExist(fullName string) bool
	IsNotExist(fullName string) bool
	IsDir(fullName string) bool
	Create(fullName string) (io.WriteCloser, error)
	Size(fullName string) int64
	ServeFile(w http.ResponseWriter, r *http.Request, name string)
	ReadFile(fullName string) ([]byte, error)
	TempFile(fullName string) ReadWriteCloser
}

// allow for impl substitutions. ie: env vars to use S3, disk otherwise.
var F = &Volume{
	FileAt: "./persistent",
}
