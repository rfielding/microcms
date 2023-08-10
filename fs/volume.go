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

func (v *Volume) At() string {
	return v.FileAt
}

func (v *Volume) Open(name string) (io.ReadCloser, error) {
	f, err := os.Open(v.FileAt + name)
	return f, err
}

func (v *Volume) ReadDir(name string) ([]fs.DirEntry, error) {
	d, err := os.ReadDir(v.FileAt + name)
	return d, err
}

func (v *Volume) Remove(name string) error {
	return os.Remove(v.FileAt + name)
}

func (v *Volume) IsExist(name string) bool {
	_, err := os.Stat(v.FileAt + name)
	return err == nil
}

func (v *Volume) IsNotExist(name string) bool {
	_, err := os.Stat(v.FileAt + name)
	return os.IsNotExist(err)
}

func (v *Volume) IsDir(name string) bool {
	s, err := os.Stat(v.FileAt + name)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func (v *Volume) Create(name string) (io.WriteCloser, error) {
	err := os.MkdirAll(path.Dir(v.FileAt+name), 0777)
	if err != nil {
		return nil, fmt.Errorf("Whil Create / MkdirAll: %v", err)
	}
	f, err := os.Create(v.FileAt + name)
	if err != nil {
		return nil, fmt.Errorf("While Create / os.Create: %v", err)
	}
	return f, err
}

func (v *Volume) MkdirAll(name string, perm fs.FileMode) error {
	return os.MkdirAll(v.FileAt+name, perm)
}

func (v *Volume) Size(name string) int64 {
	s, err := os.Stat(v.FileAt + name)
	if err != nil {
		return 0
	}
	return s.Size()
}

func (v *Volume) ServeFile(w http.ResponseWriter, r *http.Request, name string) {
	http.ServeFile(w, r, v.FileAt+name)
}

func (v *Volume) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(v.FileAt + name)
}
