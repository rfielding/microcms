package fs_test

import (
	"io"
	"os"
	"testing"

	"github.com/rfielding/microcms/fs"
)

func TestVFSSize(t *testing.T) {
	v := fs.NewVolume("../persistent")
	genericSize(t, v)
}

func TestS3VFSSize(t *testing.T) {
	v, err := fs.NewS3VFS("")
	if err != nil {
		t.Logf("could not construct S3 VFS: %v", err)
		t.Fail()
	}
	genericSize(t, v)
}

func genericSize(t *testing.T, v fs.VFS) {
	tests := []struct {
		path     string
		expected int64
	}{
		{"/files/init/defaultPermissions.rego.templ", 181},
		{"/files/init/listingTemplate.html.templ", 484},
		{"/files/init/permissions.rego", 174},
	}

	for _, test := range tests {
		result := v.Size(test.path)
		if result != test.expected {
			t.Errorf("Size(%s) = %v; want %v", test.path, result, test.expected)
		}
	}
}

func TestVFS(t *testing.T) {
	v := fs.NewVolume("../persistent")
	genericVFS(t, v)
}

func TestS3VFS(t *testing.T) {
	v, err := fs.NewS3VFS("")
	if err != nil {
		t.Logf("could not construct S3 VFS: %v", err)
		t.Fail()
	}
	genericVFS(t, v)
}

func TestIsDirS3(t *testing.T) {
	vfs, err := fs.NewS3VFS("")
	if err != nil {
		t.Fatalf("Failed to create S3VFS: %v", err)
	}
	isDirVolume(t, vfs)
}

func TestIsDirVolume(t *testing.T) {
	v := fs.NewVolume("../persistent")
	isDirVolume(t, v)
}

func isDirVolume(t *testing.T, vfs fs.VFS) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/files", true},
		{"/files/", true},
		{"/files/init/", true},
		{"/files/init", true},
		{"/files/nonexistent", false}, // We don't require paths to exist
	}

	for _, test := range tests {
		result := vfs.IsDir(test.path)
		if result != test.expected {
			t.Errorf("IsDir(%s) = %v; want %v", test.path, result, test.expected)
		}
	}
}

func TestIsExistS3(t *testing.T) {
	vfs, err := fs.NewS3VFS("")
	if err != nil {
		t.Fatalf("Failed to create S3VFS: %v", err)
	}
	isExist(t, vfs)
}

func TestIsExistVolume(t *testing.T) {
	v := fs.NewVolume("../persistent")
	isExist(t, v)
}

func isExist(t *testing.T, vfs fs.VFS) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/files", true},
		{"/files/", true},
		{"/files/init/", true},
		{"/files/init", true},
		{"/files/init0", false},
		{"/files/nonexistent", false},
	}

	for _, test := range tests {
		result := vfs.IsExist(test.path)
		if result != test.expected {
			t.Errorf("IsExist(%s) = %v; want %v", test.path, result, test.expected)
		}
	}
}

func genericVFS(t *testing.T, v fs.VFS) {
	// ASSUME that these are initialized
	if v.IsDir("/files/") == false {
		t.Logf("/files/ should be a directory")
		t.Fail()
	}
	if v.IsDir("/files/init/") == false {
		t.Logf("/files/init/ should be a directory")
		t.Fail()
	}
	if v.IsDir("/files/init") == false {
		t.Logf("/files/init should be a directory")
		t.Fail()
	}
	if v.IsDir("/files/permissions.rego") == true {
		t.Logf("/files/permissions.rego should be a file")
		t.Fail()
	}
	d, err := v.ReadDir("/files/")
	if err != nil {
		t.Logf("could read dir: %v", err)
		t.Fail()
	}
	if len(d) == 0 {
		t.Logf("/files/ should have at least one file in it")
		t.Fail()
	}
	d, err = v.ReadDir("/files")
	if err != nil {
		t.Logf("could read dir: %v", err)
		t.Fail()
	}
	if len(d) == 0 {
		t.Logf("/files should have at least one file in it")
		t.Fail()
	}
	for i := range d {
		isDir := d[i].IsDir()
		dirMark := "[f]"
		if isDir {
			dirMark = "[d]"
		}
		t.Logf("  %s %s", dirMark, d[i].Name())
	}
	indexhtml, err := v.Open("/files/init/rootTemplate.html.templ")
	if err != nil {
		t.Logf("could read /files/init/rootTemplate.html.templ: %v", err)
		t.Fail()
	}
	defer indexhtml.Close()
	io.Copy(os.Stdout, indexhtml)
}
