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

// Handling At() implies that the local conversion programs may move into here,
// because they use `name`

type VFS interface {
	At() string // hmm. still have filesystem dependencies due to ImageMagick, etc.
	Open(name string) (io.ReadCloser, error)
	ReadDir(name string) ([]fs.DirEntry, error)
	Remove(name string) error
	IsExist(name string) bool
	IsNotExist(name string) bool
	IsDir(name string) bool
	Create(name string) (io.WriteCloser, error)
	Size(name string) int64
	ServeFile(w http.ResponseWriter, r *http.Request, name string)
	ReadFile(name string) ([]byte, error)
}

// allow for impl substitutions
var F = &Volume{
	FileAt: "./persistent",
}
