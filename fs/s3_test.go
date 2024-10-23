package fs_test

import (
	"io"
	"os"
	"testing"

	"github.com/rfielding/microcms/fs"
)

// Test ReadDir entries
func TestReadDirS3(t *testing.T) {
	vfs, err := fs.NewS3VFS("")
	if err != nil {
		t.Fatalf("Failed to create S3VFS: %v", err)
	}
	readDir(t, vfs)
}

func TestReadDirVolume(t *testing.T) {
	v := fs.NewVolume("../persistent")
	readDir(t, v)
}

/*
match this data:

> ls -al files/init
total 36
drwxr-xr-x 3 rfielding rfielding 4096 Jul  2 22:58 .
drwxr-xr-x 5 rfielding rfielding 4096 Jul  2 22:59 ..
-rw-r--r-- 1 rfielding rfielding  181 Jul  2 22:58 defaultPermissions.rego.templ
-rw-r--r-- 1 rfielding rfielding  484 Jul  2 22:58 listingTemplate.html.templ
-rw-r--r-- 1 rfielding rfielding  174 Jul  2 22:58 permissions.rego
-rw-r--r-- 1 rfielding rfielding  392 Jul  2 22:58 rootTemplate.html.templ
-rw-r--r-- 1 rfielding rfielding  599 Jul  2 22:58 searchTemplate.html.templ
-rw-r--r-- 1 rfielding rfielding  485 Jul  2 22:58 styles.css
drwxr-xr-x 3 rfielding rfielding 4096 Jul  2 22:58 ui
*/
func readDir(t *testing.T, vfs fs.VFS) {
	// test ReadDir entries for correct file sizes
	tests := map[string]int64{
		"/files/init/defaultPermissions.rego.templ": 181,
		"/files/init/listingTemplate.html.templ":    484,
		"/files/init/permissions.rego":              174,
		"/files/init/rootTemplate.html.templ":       392,
		"/files/init/searchTemplate.html.templ":     599,
		"/files/init/styles.css":                    485,
	}
	// Test against the vfs.Size function
	for path, size := range tests {
		result := vfs.Size(path)
		if result != size {
			t.Errorf("Size(%s) = %v; want %v", path, result, size)
		}
	}
	// Now, use ReadDir to get the directory entries
	// and test those sizes.
	d, err := vfs.ReadDir("/files/init/")
	if err != nil {
		t.Logf("could not read dir: %v", err)
		t.Fail()
	}
	if len(d) == 0 {
		t.Logf("/files/init/ should have at least one file in it")
		t.Fail()
	}
	for _, entry := range d {
		path := "/files/init/" + entry.Name()
		size := tests[path]
		info, err := entry.Info()
		if err != nil {
			t.Logf("could not get info: %v", err)
			t.Fail()
		}
		if info == nil {
			t.Logf("info is nil")
			t.Fail()
		}
		if info.IsDir() {
			if info.Name() != "ui" {
				t.Errorf("IsDir(%s) = %v; want %v", path, info.Name(), "ui")
			}
		} else {
			if false && info.Size() != size {
				t.Errorf("Size(%s) = %v; want %v", path, info.Size(), size)
			}
		}
	}
}

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
