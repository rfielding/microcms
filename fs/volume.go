package fs

import (
	"fmt"
	"io"
	"io/fs"
	"log"
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

func (v *Volume) Open(fullName string) (io.ReadCloser, error) {
	f, err := os.Open(v.FileAt + fullName)
	return f, err
}

func (v *Volume) ReadDir(fullName string) ([]fs.DirEntry, error) {
	d, err := os.ReadDir(v.FileAt + fullName)
	return d, err
}

func (v *Volume) Remove(fullName string) error {
	return os.Remove(v.FileAt + fullName)
}

func (v *Volume) IsExist(fullName string) bool {
	_, err := os.Stat(v.FileAt + fullName)
	return err == nil
}

func (v *Volume) IsNotExist(fullName string) bool {
	_, err := os.Stat(v.FileAt + fullName)
	return os.IsNotExist(err)
}

func (v *Volume) IsDir(fullName string) bool {
	s, err := os.Stat(v.FileAt + fullName)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func (v *Volume) Create(fullName string) (io.WriteCloser, error) {
	err := os.MkdirAll(path.Dir(v.FileAt+fullName), 0777)
	if err != nil {
		return nil, fmt.Errorf("Whil Create / MkdirAll: %v", err)
	}
	f, err := os.Create(v.FileAt + fullName)
	if err != nil {
		return nil, fmt.Errorf("While Create / os.Create: %v", err)
	}
	return f, err
}

func (v *Volume) MkdirAll(fullName string, perm fs.FileMode) error {
	return os.MkdirAll(v.FileAt+fullName, perm)
}

func (v *Volume) Size(fullName string) int64 {
	s, err := os.Stat(v.FileAt + fullName)
	if err != nil {
		return 0
	}
	return s.Size()
}

func (v *Volume) ServeFile(w http.ResponseWriter, r *http.Request, name string) {
	http.ServeFile(w, r, v.FileAt+name)
}

func (v *Volume) ReadFile(fullName string) ([]byte, error) {
	return os.ReadFile(v.FileAt + fullName)
}

func (v *Volume) TempFile(fullName string) (ReadWriteCloser, error) {
	file, err := os.CreateTemp(v.FileAt, "microcms-*-.tmp")
	if err != nil {
		log.Fatal(err)
	}
	return file, err
}
