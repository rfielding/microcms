package handler

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/rfielding/microcms/data"
	"github.com/rfielding/microcms/fs"
)

type LabelModel struct {
	LabelModelVersion string        `json:"LabelModelVersion"`
	Labels            []LabelObject `json:"Labels"`
}

type LabelObject struct {
	Name string `json:"Name"`
}

// When we return the error type, also suggest the http error to return,
// as at the time the error happens, that is when it is best known how to handle it.
type HttpError int

func RecompileTemplates(fullName string) {
	if fullName == "/files/init/searchTemplate.html.templ" {
		log.Printf("Recompiling %s", fullName)
		compiledSearchTemplate = compileTemplate(fullName)
	}
	if fullName == "/files/init/listingTemplate.html.templ" {
		log.Printf("Recompiling %s", fullName)
		compiledListingTemplate = compileTemplate(fullName)
	}
	if fullName == "/files/init/rootTemplate.html.templ" {
		log.Printf("Recompiling %s", fullName)
		compiledRootTemplate = compileTemplate(fullName)
	}
	if fullName == "/files/init/defaultPermissions.rego.templ" {
		log.Printf("Recompiling %s", fullName)
		compiledDefaultPermissionsTemplate = compileTemplate(fullName)
	}
}

func postFileHandler(
	user data.User,
	stream io.Reader,
	fsPath string,
	fsName string,
	originalFsPath string, // the path that triggered the creation of this file
	originalFsName string, // the file that triggered the creation of this file
	cascade bool, // allow further derivatives
	privileged bool, // ignore the permissions, because this is startup
) (HttpError, error) {
	// XXX: there is some issue with trailing slashes to make an fsPath work correctly.
	// when that is correct, it will be fullName := fsPath + fsName
	// the name fsPath implies that it is a known directory with a trailing slash.
	//    fsPath == path.Dir(fullName) + "/"

	if !strings.HasSuffix(fsPath, "/") {
		log.Printf("!!!! inconsistent fsPath ends in slash here! %s", fsPath+fsName)
	}
	if !strings.HasSuffix(originalFsPath, "/") {
		log.Printf("!!!! inconsistent originalFsPath ends in slash here! %s", fsPath+fsName)
	}

	fullName := fsPath + fsName

	if !privileged && !CanWrite(user, fsPath, fsName) {
		return http.StatusForbidden, fmt.Errorf(
			"POST disallowed on %s for %s",
			fullName, UserName(user),
		)
	}

	// reject requests that would get this user stuck in fixing it
	if fsName == "permissions.rego" || strings.HasSuffix(fsName, "--permissions.rego") {
		proposedUpload, err := ioutil.ReadAll(stream)
		if err != nil {
			return http.StatusForbidden, fmt.Errorf(
				"Could not read proposed permissions.rego: %v",
				err,
			)
		}
		proposedAttrs, err := CalculateRego(user, string(proposedUpload))
		if err != nil {
			return http.StatusForbidden, fmt.Errorf(
				"Could not calculate proposed permissions.rego: %v",
				err,
			)
		}
		if !proposedAttrs.Write || !proposedAttrs.Read {
			return http.StatusForbidden, fmt.Errorf(
				"Proposed permissions.rego does not allow write and read: %v",
				proposedAttrs,
			)
		}
		stream = bytes.NewReader(proposedUpload)
	}

	// Ensure that the file in question exists on disk.
	if true {
		f, err := fs.F.Create(fullName)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("Could not create file %s: %v", fullName, err)
		}

		// Save the stream to a file
		sz, err := io.Copy(f, stream)
		f.Close() // strange positioning, but we must close before defer can get to it.
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("Could not write to file (%d bytes written) %s: %v", sz, fullName, err)
		}

		// Make sure that these are re-compiled on upload
		RecompileTemplates(fullName)
	}

	if IsDoc(fullName) && cascade {
		// Open the file we wrote
		f, err := fs.F.Open(fullName)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("Could not open file for indexing %s: %v", fullName, err)
		}
		// Get a doc extract stream
		rdr, err := DocExtract(fullName, f)
		f.Close()
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("Could not extract file for indexing %s: %v", fullName, err)
		}
		// Write the doc extract stream like an upload
		extractName := fmt.Sprintf("%s--extract.txt", fsName)
		herr, err := postFileHandler(user, rdr, fsPath, extractName, originalFsPath, originalFsName, cascade, privileged)
		if err != nil {
			return herr, fmt.Errorf("Could not write extract file for indexing %s: %v", fullName, err)
		}

		ext := strings.ToLower(path.Ext(fullName))
		if ext == ".pdf" {
			rdr, err := pdfThumbnail(fullName)
			if err != nil {
				return http.StatusInternalServerError, fmt.Errorf("Could not make thumbnail for %s: %v", fullName, err)
			}
			// Only png works.  bug in imageMagick.  don't cascade on thumbnails
			thumbnailName := fmt.Sprintf("%s--thumbnail.png", fsName)
			herr, err := postFileHandler(user, rdr, fsPath, thumbnailName, originalFsPath, originalFsName, false, privileged)
			if err != nil {
				return herr, fmt.Errorf("Could not write make thumbnail for indexing %s: %v", fullName, err)
			}
		}

		// open the file that we saved, and index it in the database.
		return http.StatusOK, nil
	}

	if IsVideo(fullName) && cascade {
		rdr, err := videoThumbnail(fullName)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("Could not make thumbnail for %s: %v", fullName, err)
		}
		thumbnailName := fmt.Sprintf("%s--thumbnail.png", fsName)
		herr, err := postFileHandler(user, rdr, fsPath, thumbnailName, originalFsPath, originalFsName, false, privileged)
		if err != nil {
			return herr, fmt.Errorf("Could not write make thumbnail for indexing %s: %v", fullName, err)
		}
		return http.StatusOK, nil
	}

	if IsImage(fullName) && cascade {
		if true {
			rdr, err := MakeThumbnail(fullName)
			if err != nil {
				return http.StatusInternalServerError, fmt.Errorf("Could not make thumbnail for %s: %v", fullName, err)
			}
			thumbnailName := fmt.Sprintf("%s--thumbnail.png", fsName)
			herr, err := postFileHandler(user, rdr, fsPath, thumbnailName, originalFsPath, originalFsName, false, privileged)
			if err != nil {
				return herr, fmt.Errorf("Could not write make thumbnail for indexing %s: %v", fullName, err)
			}
		}

		if os.Getenv("AWS_ACCESS_KEY_ID") != "" {
			{
				log.Printf("detect labels on %s", fullName)
				rdr, err := detectLabels(fullName)
				if err != nil {
					return http.StatusInternalServerError, fmt.Errorf("Could not extract labels for %s: %v", fullName, err)
				}
				labelName := fmt.Sprintf("%s--labels.json", fsName)
				herr, err := postFileHandler(user, rdr, fsPath, labelName, originalFsPath, originalFsName, cascade, privileged)
				if err != nil {
					return herr, fmt.Errorf("Could not write labgel detect %s: %v", fullName, err)
				}
				// re-read full file off of disk. TODO: maybe better to parse and pass json to avoid it
				labelFile := fullName + "--labels.json"
				jf, err := fs.F.ReadFile(labelFile)
				if err != nil {
					return http.StatusInternalServerError, fmt.Errorf("Could not find file: %s %v", labelFile, err)
				}
				var j LabelModel
				err = json.Unmarshal(jf, &j)
				if err != nil {
					return http.StatusInternalServerError, fmt.Errorf("Could not look for celeb detect on labels for %s", fullName)
				} else {
					for i := range j.Labels {
						v := j.Labels[i].Name
						if v == "Face" || v == "Person" || v == "People" {
							log.Printf("detect celebs on %s", fullName)
							rdr, err = detectCeleb(fullName)
							if err != nil {
								return http.StatusInternalServerError, fmt.Errorf("Could not extract labels for %s: %v", fullName, err)
							}
							if rdr != nil {
								faceName := fmt.Sprintf("%s--celebs.json", fsName)
								herr, err := postFileHandler(user, rdr, fsPath, faceName, originalFsPath, originalFsName, cascade, privileged)
								if err != nil {
									return herr, fmt.Errorf("Could not write face detect %s: %v", fullName, err)
								}
								break
							}
						}
					}
				}
			}
			{
				log.Printf("content moderation on %s", fullName)
				rdr, err := detectModeration(fullName)
				if err != nil {
					return http.StatusInternalServerError, fmt.Errorf("Could not do content moderation for %s: %v", fullName, err)
				}
				labelName := fmt.Sprintf("%s--moderation.json", fsName)
				herr, err := postFileHandler(user, rdr, fsPath, labelName, originalFsPath, originalFsName, cascade, privileged)
				if err != nil {
					return herr, fmt.Errorf("Could not content moderate %s: %v", fullName, err)
				}
			}
		}

		return http.StatusOK, nil
	}

	if IsTextFile(fullName) && cascade {
		// open the file that we saved, and index it in the database.
		f, err := fs.F.Open(fullName)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("Could not open file for indexing %s: %v", fullName, err)
		}
		defer f.Close()

		var rdr io.Reader = f
		// chunk sizes for making search results
		buffer := make([]byte, 16*1024)
		part := 0
		for {
			sz, err := rdr.Read(buffer)
			if err == io.EOF {
				break
			}
			err = indexTextFile(fsPath, fsName, part, originalFsPath, originalFsName, buffer[:sz])
			if err != nil {
				return http.StatusInternalServerError, fmt.Errorf("failed indexing: %v", err)
			}
			part++
		}
		return http.StatusOK, nil
	}
	return http.StatusOK, nil
}

func postFilesHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) {
	defer r.Body.Close()
	user := data.GetUser(r)

	if ensureThatHomeDirExists(w, r, user) {
		return
	}

	if cl := r.Header.Get("Content-Length"); cl != "" {
		n, err := strconv.Atoi(cl)
		if err == nil {
			t := MetricsGet.Task()
			t.BytesRead += int64(n)
			defer t.End()
		}
	}

	q := r.URL.Query()
	// This is a signal that this is a tar archive
	// that we unpack to install all files at the given url
	needsInstall := q.Get("install") == "true"
	if needsInstall {
		log.Printf("install tarball to %s", r.URL.Path)
	}

	// Make sure that the path exists, and get the file name
	parentDir := strings.Join(pathTokens[:len(pathTokens)-1], "/")
	fsPath := parentDir + "/"
	fsName := pathTokens[len(pathTokens)-1]

	// If err != nil, then we can't call this again.  http response has been sent
	if needsInstall == true { // we ASSUME that it's a tar
		// We install the tarball, walking it file by file
		t := tar.NewReader(r.Body)
		for {
			header, err := t.Next()
			if err == io.EOF {
				break
			}
			// Ignore directories for a moment XXX
			if header.Typeflag == tar.TypeReg {
				// I assume that header names are unqualified dir names
				tname := strings.Split(header.Name, "/") // expect . in front
				tname = tname[1:]
				tarFsDir := path.Dir(fmt.Sprintf("%s/%s/%s", parentDir, fsName, strings.Join(tname, "/"))) + "/"
				tarFsName := path.Base(header.Name)
				log.Printf("writing: %s into %s", tarFsName, tarFsDir)
				herr, err := postFileHandler(user, t, tarFsDir, tarFsName, tarFsDir, tarFsName, true, false)
				if err != nil {
					w.WriteHeader(int(herr))
					w.Write([]byte(fmt.Sprintf("ERR: %v", err)))
					log.Printf("ERR %v", err)
					return
				}
			}
		}
	} else {
		// Just a normal single-file upload
		herr, err := postFileHandler(user, r.Body, fsPath, fsName, fsPath, fsName, true, false)
		if err != nil {
			w.WriteHeader(int(herr))
			w.Write([]byte(fmt.Sprintf("ERR: %v", err)))
			log.Printf("ERR %v", err)
			return
		}
	}
}
