package handler

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/rfielding/microcms/data"
	"github.com/rfielding/microcms/fs"
)

func deleteHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) {
	user := data.GetUser(r)
	if !CanWrite(user, path.Dir(r.URL.Path), path.Base(r.URL.Path)) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(fmt.Sprintf(
			"DELETE disallowed on %s for %s", r.URL.Path, user.Identity(),
		)))
		return
	}
	if strings.HasPrefix(r.URL.Path, "/files/") {
		// If this would break our access, then don't allow it
		fsName := path.Base(r.URL.Path)
		if fsName == "permissions.rego" || strings.HasSuffix(fsName, "--permissions.rego") {
			parent := path.Base(path.Dir(r.URL.Path))
			grandparent := path.Dir(path.Dir(r.URL.Path)) + "/"
			attrsAfter := GetAttrs(user, grandparent, parent)
			if !attrsAfter.Write || !attrsAfter.Read {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(fmt.Sprintf(
					"DELETE disallowed on %s for %s", r.URL.Path, user.Identity(),
				)))
				return
			}
		}

		if fs.F.IsExist(r.URL.Path) {
			t := MetricsDelete.Task()
			t.BytesWrite += fs.F.Size(r.URL.Path)
			defer t.End()
			err := fs.F.Remove(r.URL.Path)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("unable to delete file: %v", err)))
				return
			}
		}
		// If it's gone, then it's deleted
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
