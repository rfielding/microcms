package fs

import (
	"io/fs"
	"net/http"
	"os"
)

/*
 Wrap up os, so that we dont have to make physical paths from logical paths,
 and to remove all the raw stat calls.
*/

func Open(name string) (*os.File, error) {
	f, err := os.Open(name)
	return f, err
}

func ReadDir(name string) ([]fs.DirEntry, error) {
	d, err := os.ReadDir(name)
	return d, err
}

func IsExist(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func IsNotExist(name string) bool {
	_, err := os.Stat(name)
	return os.IsNotExist(err)
}

func IsDir(name string) bool {
	s, err := os.Stat(name)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func Create(name string) (*os.File, error) {
	f, err := os.Create(name)
	return f, err
}

func MkdirAll(name string, perm fs.FileMode) error {
	return os.MkdirAll(name, perm)
}

func Size(name string) int64 {
	s, err := os.Stat(name)
	if err != nil {
		return 0
	}
	return s.Size()
}

func ServeFile(w http.ResponseWriter, r *http.Request, name string) {
	http.ServeFile(w, r, name)
}
