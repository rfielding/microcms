package fs_test

import (
 "os"
 "io"
 "testing"
 "github.com/rfielding/microcms/fs"
)

func TestVFS(t *testing.T) {
	v := fs.NewVolume("../persistent")
	genericVFS(t,v)
}

func xTestS3VFS(t *testing.T) {
	v, err := fs.NewS3VFS("")
	if err != nil {
		t.Logf("could not construct S3 VFS: %v", err)
		t.Fail()
	}
	genericVFS(t,v)
}

func xTestIsDirS3(t *testing.T) {
    vfs, err := fs.NewS3VFS("")
    if err != nil {
        t.Fatalf("Failed to create S3VFS: %v", err)
    }
    isDirVolume(t, vfs)
}

func TestIsDirVolume(t *testing.T) {
	v := fs.NewVolume("../persistent")
	isDirVolume(t,v)
}

func isDirVolume(t *testing.T, vfs fs.VFS) {
    tests := []struct {
        path     string
        expected bool
    }{
        {"/files/", true},
        {"/files/init/", true},
        {"/files/init", true}, // This might depend on your bucket structure
        {"/files/nonexistent", false},
    }

    for _, test := range tests {
        result := vfs.IsDir(test.path)
        if result != test.expected {
            t.Errorf("IsDir(%s) = %v; want %v", test.path, result, test.expected)
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
	io.Copy(os.Stdout,indexhtml)	
}

