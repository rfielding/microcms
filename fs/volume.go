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
	at string
	fs http.Handler
}

func NewVolume(at string) *Volume {
	fa := at
	return &Volume{
		fs: http.FileServer(http.Dir(fa)),
		at: fa,
	}
}

func (v *Volume) Open(fullName string) (io.ReadCloser, error) {
	f, err := os.Open(v.at + fullName)
	return f, err
}

func (v *Volume) ReadDir(fullName string) ([]fs.DirEntry, error) {
	d, err := os.ReadDir(v.at + fullName)
	return d, err
}

func (v *Volume) Remove(fullName string) error {
	return os.Remove(v.at + fullName)
}

func (v *Volume) IsExist(fullName string) bool {
	_, err := os.Stat(v.at + fullName)
	return err == nil
}

func (v *Volume) IsNotExist(fullName string) bool {
	_, err := os.Stat(v.at + fullName)
	return os.IsNotExist(err)
}

func (v *Volume) IsDir(fullName string) bool {
	s, err := os.Stat(v.at + fullName)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func (v *Volume) Create(fullName string) (io.WriteCloser, error) {
	err := os.MkdirAll(path.Dir(v.at+fullName), 0777)
	if err != nil {
		return nil, fmt.Errorf("Whil Create / MkdirAll: %v", err)
	}
	f, err := os.Create(v.at + fullName)
	if err != nil {
		return nil, fmt.Errorf("While Create / os.Create: %v", err)
	}
	return f, err
}

func (v *Volume) MkdirAll(fullName string, perm fs.FileMode) error {
	return os.MkdirAll(v.at+fullName, perm)
}

func (v *Volume) Size(fullName string) int64 {
	s, err := os.Stat(v.at + fullName)
	if err != nil {
		return 0
	}
	return s.Size()
}

func (v *Volume) ServeFile(w http.ResponseWriter, r *http.Request, fullName string) {
	http.ServeFile(w, r, v.at+fullName)
}

func (v *Volume) ReadFile(fullName string) ([]byte, error) {
	return os.ReadFile(v.at + fullName)
}

func (v *Volume) TempFile(fullName string) (ReadWriteCloser, error) {
	file, err := os.CreateTemp(v.at, "microcms-*-.tmp")
	if err != nil {
		log.Fatal(err)
	}
	return file, err
}

func (v *Volume) FileServer() http.Handler {
	return v.fs
}

func (v *Volume) PdfThumbnail(fullName string) (io.Reader, error) {
	command := []string{
		"convert",
		"-resize", "x100",
		(v.at + fullName + "[0]"),
		"png:-",
	}
	return commandReader(command)
}

func (v *Volume) MakeThumbnail(fullName string) (io.Reader, error) {
	command := []string{
		"convert",
		"-thumbnail", "x100",
		"-background", "white",
		"-alpha", "remove",
		"-format", "png",
		(v.at + fullName),
		"-",
	}
	return commandReader(command)
}

func (v *Volume) VideoThumbnail(fullName string) (io.Reader, error) {
	command := []string{
		"convert",
		"-resize", "x100",
		(v.at + fullName + "[100]"),
		"png:-",
	}
	return commandReader(command)
}
