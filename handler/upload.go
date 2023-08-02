package handler

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
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

// postFileHandler can be re-used as long as err != nil
func postFileHandler(
	user data.User,
	stream io.Reader,
	parentDir string,
	name string,
	originalParentDir string, // the path that triggered the creation of this file
	originalName string, // the file that triggered the creation of this file
	cascade bool, // allow further derivatives
	privileged bool, // ignore the permissions, because this is startup
) (HttpError, error) {

	if !privileged && !CanWrite(user, parentDir, name) {
		return http.StatusForbidden, fmt.Errorf("write disallowed")
	}

	fullName := fmt.Sprintf("%s/%s", parentDir, name)

	//log.Printf("Ensure existence of parentDir: %s", parentDir)
	err := fs.MkdirAll(parentDir, 0777)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Could not create path for %s: %v", parentDir, err)
	}

	existingSize := fs.Size(fullName)

	// Ensure that the file in question exists on disk.
	if true {
		f, err := fs.Create(fullName)
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

	if IsDoc(fullName) && cascade {
		// Open the file we wrote
		f, err := fs.Open(fullName)
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
		extractName := fmt.Sprintf("%s--extract.txt", name)
		herr, err := postFileHandler(user, rdr, parentDir, extractName, originalParentDir, originalName, cascade, privileged)
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
			thumbnailName := fmt.Sprintf("%s--thumbnail.png", name)
			herr, err := postFileHandler(user, rdr, parentDir, thumbnailName, originalParentDir, originalName, false, privileged)
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
		thumbnailName := fmt.Sprintf("%s--thumbnail.png", name)
		herr, err := postFileHandler(user, rdr, parentDir, thumbnailName, originalParentDir, originalName, false, privileged)
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
			thumbnailName := fmt.Sprintf("%s--thumbnail.png", name)
			herr, err := postFileHandler(user, rdr, parentDir, thumbnailName, originalParentDir, originalName, false, privileged)
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
				labelName := fmt.Sprintf("%s--labels.json", name)
				herr, err := postFileHandler(user, rdr, parentDir, labelName, originalParentDir, originalName, cascade, privileged)
				if err != nil {
					return herr, fmt.Errorf("Could not write labgel detect %s: %v", fullName, err)
				}
				// re-read full file off of disk. TODO: maybe better to parse and pass json to avoid it
				labelFile := fullName + "--labels.json"
				jf, err := fs.ReadFile(labelFile)
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
								faceName := fmt.Sprintf("%s--celebs.json", name)
								herr, err := postFileHandler(user, rdr, parentDir, faceName, originalParentDir, originalName, cascade, privileged)
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
				labelName := fmt.Sprintf("%s--moderation.json", name)
				herr, err := postFileHandler(user, rdr, parentDir, labelName, originalParentDir, originalName, cascade, privileged)
				if err != nil {
					return herr, fmt.Errorf("Could not content moderate %s: %v", fullName, err)
				}
			}
		}

		return http.StatusOK, nil
	}

	if IsTextFile(fullName) && cascade {
		// open the file that we saved, and index it in the database.
		f, err := fs.Open(fullName)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("Could not open file for indexing %s: %v", fullName, err)
		}
		defer f.Close()
		if existingSize > 0 {
			// we are appending, so we need to start at the end of the file
			f.Seek(existingSize, 0)
		}
		var rdr io.Reader = f
		buffer := make([]byte, 4*1024)
		part := 0
		for {
			sz, err := rdr.Read(buffer)
			if err == io.EOF {
				break
			}
			err = indexTextFile(parentDir+"/", name, part, originalParentDir+"/", originalName, buffer[:sz])
			if err != nil {
				// just let it go and continue??
				return http.StatusInternalServerError, fmt.Errorf("failed indexing: %v", err)
			}
			part++
		}
		return http.StatusOK, nil
	}
	return http.StatusOK, nil
}

func postFilesHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) {
	var err error
	defer r.Body.Close()
	user := data.GetUser(r)

	if len(pathTokens) < 2 {
		err := fmt.Errorf("path needs /[command]/[url] for post to %s: %v", r.URL.Path, err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("ERR: %v", err)))
		log.Printf("ERR %v", err)
		return
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
	name := pathTokens[len(pathTokens)-1]

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
				tardir := path.Dir(fmt.Sprintf("%s/%s/%s", parentDir, name, strings.Join(tname, "/")))
				tarname := path.Base(header.Name)
				log.Printf("writing: %s into %s", tarname, tardir)
				herr, err := postFileHandler(user, t, tardir, tarname, tardir, tarname, true, false)
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
		herr, err := postFileHandler(user, r.Body, parentDir, name, parentDir, name, true, false)
		if err != nil {
			w.WriteHeader(int(herr))
			w.Write([]byte(fmt.Sprintf("ERR: %v", err)))
			log.Printf("ERR %v", err)
			return
		}
	}
}
