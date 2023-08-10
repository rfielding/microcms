package fs

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
)

func (fsi *Fs) Open(name string) (io.ReadCloser, error) {
	f, err := os.Open(At + name)
	return f, err
}

func (fsi *Fs) ReadDir(name string) ([]fs.DirEntry, error) {
	d, err := os.ReadDir(At + name)
	return d, err
}

func (fsi *Fs) Remove(name string) error {
	return os.Remove(name)
}

func (fsi *Fs) IsExist(name string) bool {
	_, err := os.Stat(At + name)
	return err == nil
}

func (fsi *Fs) IsNotExist(name string) bool {
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

func (fsi *Fs) Create(name string) (io.WriteCloser, error) {
	err := os.MkdirAll(path.Base(At+name), 0777)
	if err != nil {
		return nil, err
	}
	f, err := os.Create(At + name)
	return f, err
}

func (fsi *Fs) MkdirAll(name string, perm fs.FileMode) error {
	return os.MkdirAll(At+name, perm)
}

func (fsi *Fs) Size(name string) int64 {
	s, err := os.Stat(At + name)
	if err != nil {
		return 0
	}
	return s.Size()
}

func (fsi *Fs) ServeFile(w http.ResponseWriter, r *http.Request, name string) {
	http.ServeFile(w, r, At+name)
}

func (fsi *Fs) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(At + name)
}
