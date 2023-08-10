package fs

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
)

type Volume struct {
	FileAt string
}

func (fsi *Volume) At() string {
	return fsi.FileAt
}

func (fsi *Volume) Open(name string) (io.ReadCloser, error) {
	f, err := os.Open(fsi.At() + name)
	return f, err
}

func (fsi *Volume) ReadDir(name string) ([]fs.DirEntry, error) {
	d, err := os.ReadDir(fsi.At() + name)
	return d, err
}

func (fsi *Volume) Remove(name string) error {
	return os.Remove(fsi.At() + name)
}

func (fsi *Volume) IsExist(name string) bool {
	_, err := os.Stat(fsi.At() + name)
	return err == nil
}

func (fsi *Volume) IsNotExist(name string) bool {
	_, err := os.Stat(fsi.At() + name)
	return os.IsNotExist(err)
}

func (fsi *Volume) IsDir(name string) bool {
	s, err := os.Stat(fsi.At() + name)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func (fsi *Volume) Create(name string) (io.WriteCloser, error) {
	err := os.MkdirAll(path.Dir(fsi.At()+name), 0777)
	if err != nil {
		return nil, fmt.Errorf("Whil Create / MkdirAll: %v", err)
	}
	f, err := os.Create(fsi.At() + name)
	if err != nil {
		return nil, fmt.Errorf("While Create / os.Create: %v", err)
	}
	return f, err
}

func (fsi *Volume) MkdirAll(name string, perm fs.FileMode) error {
	return os.MkdirAll(fsi.At()+name, perm)
}

func (fsi *Volume) Size(name string) int64 {
	s, err := os.Stat(fsi.At() + name)
	if err != nil {
		return 0
	}
	return s.Size()
}

func (fsi *Volume) ServeFile(w http.ResponseWriter, r *http.Request, name string) {
	http.ServeFile(w, r, fsi.At()+name)
}

func (fsi *Volume) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(fsi.At() + name)
}
