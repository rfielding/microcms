package fs

import (
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
)

/*
 Wrap up os, so that we dont have to make physical paths from logical paths,
 and to remove all the raw stat calls.
*/

const At = "."

func Open(name string) (*os.File, error) {
	f, err := os.Open(At + name)
	return f, err
}

func ReadDir(name string) ([]fs.FileInfo, error) {
	d, err := ioutil.ReadDir(At + name)
	return d, err
}

func IsExist(name string) bool {
	_, err := os.Stat(At + name)
	return err == nil
}

func IsNotExist(name string) bool {
	_, err := os.Stat(At + name)
	return os.IsNotExist(err)
}

func IsDir(name string) bool {
	s, err := os.Stat(At + name)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func Create(name string) (*os.File, error) {
	f, err := os.Create(At + name)
	return f, err
}

func MkdirAll(name string, perm fs.FileMode) error {
	return os.MkdirAll(At+name, perm)
}

func Size(name string) int64 {
	s, err := os.Stat(At + name)
	if err != nil {
		return 0
	}
	return s.Size()
}

func ServeFile(w http.ResponseWriter, r *http.Request, name string) {
	http.ServeFile(w, r, At+name)
}

func ReadFile(name string) ([]byte, error) {
	return ioutil.ReadFile(At + name)
}
