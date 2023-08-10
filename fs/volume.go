package fs

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
)

type Volume struct{}

func (fsi *Volume) Open(name string) (io.ReadCloser, error) {
	f, err := os.Open(At + name)
	return f, err
}

func (fsi *Volume) ReadDir(name string) ([]fs.DirEntry, error) {
	d, err := os.ReadDir(At + name)
	return d, err
}

func (fsi *Volume) Remove(name string) error {
	return os.Remove(name)
}

func (fsi *Volume) IsExist(name string) bool {
	_, err := os.Stat(At + name)
	return err == nil
}

func (fsi *Volume) IsNotExist(name string) bool {
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

func (fsi *Volume) Create(name string) (io.WriteCloser, error) {
	err := os.MkdirAll(path.Base(At+name), 0777)
	if err != nil {
		return nil, fmt.Errorf("Whil Create / MkdirAll: %v", err)
	}
	f, err := os.Create(At + name)
	if err != nil {
		return nil, fmt.Errorf("While Create / os.Create: %v", err)
	}
	return f, err
}

func (fsi *Volume) MkdirAll(name string, perm fs.FileMode) error {
	return os.MkdirAll(At+name, perm)
}

func (fsi *Volume) Size(name string) int64 {
	s, err := os.Stat(At + name)
	if err != nil {
		return 0
	}
	return s.Size()
}

func (fsi *Volume) ServeFile(w http.ResponseWriter, r *http.Request, name string) {
	http.ServeFile(w, r, At+name)
}

func (fsi *Volume) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(At + name)
}
