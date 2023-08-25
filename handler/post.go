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

func postHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) {
	if len(pathTokens) > 2 && pathTokens[1] == "files" {
		postFilesHandler(w, r, pathTokens)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

var DocExtractor string

// Make a request to tika in this case
func DocExtract(fullName string, rdr io.Reader) (io.ReadCloser, error) {
	cl := http.Client{}
	req, err := http.NewRequest("PUT", DocExtractor, rdr)
	if err != nil {
		return nil, fmt.Errorf("Unable to make request to upload file: %v", err)
	}
	req.Header.Add("accept", "text/plain")
	res, err := cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Unable to do request to upload file %s: %v", fullName, err)
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Unable to upload %s: %d", fullName, res.StatusCode)
	}
	return res.Body, nil
}

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

func recompileTemplates(fullName string) {
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
			fullName, user.Identity(),
		)
	}

	// reject requests that would get this user stuck in fixing it
	stream, shouldReturn, returnValue, returnValue1 := badUploadPermissions(fsName, stream, user)
	if shouldReturn {
		return returnValue, returnValue1
	}

	// Ensure that the file in question exists on disk.
	// Save the stream to a file
	// strange positioning, but we must close before defer can get to it.
	// Make sure that these are re-compiled on upload
	shouldReturn1, returnValue2, returnValue3 := putFileOnDisk(fullName, stream)
	if shouldReturn1 {
		return returnValue2, returnValue3
	}

	shouldReturn, returnValue, err := IndexFileName(cascade, fsPath, fsName)
	if shouldReturn {
		return returnValue, err
	}

	// Open the file we wrote
	// Get a doc extract stream
	// Write the doc extract stream like an upload
	// Only png works.  bug in imageMagick.  don't cascade on thumbnails
	// open the file that we saved, and index it in the database.
	shouldReturn2, returnValue4, returnValue5 := extractDoc(cascade, fsName, user, fsPath, originalFsPath, originalFsName, privileged)
	if shouldReturn2 {
		return returnValue4, returnValue5
	}

	// it could easily fail on account of size, so just ignore it if it fails
	// so, ignore it until I have a better solution.
	if false {
		shouldReturn3, returnValue6, returnValue7 := extractVideo(cascade, fsName, user, fsPath, originalFsPath, originalFsName, privileged)
		if shouldReturn3 {
			return returnValue6, returnValue7
		}
	}

	// re-read full file off of disk. TODO: maybe better to parse and pass json to avoid it
	shouldReturn4, returnValue8, returnValue9 := extractImage(cascade, fsName, user, fsPath, originalFsPath, originalFsName, privileged)
	if shouldReturn4 {
		return returnValue8, returnValue9
	}

	// open the file that we saved, and index it in the database.
	// chunk sizes for making search results
	shouldReturn5, returnValue10, returnValue11 := indexPlaintext(cascade, fsPath, fsName, originalFsPath, originalFsName)
	if shouldReturn5 {
		return returnValue10, returnValue11
	}
	return http.StatusOK, nil
}

func IndexFileName(cascade bool, fsPath string, fsName string) (bool, HttpError, error) {
	// part 0 is trying to index the filename
	if cascade && !strings.Contains(fsName, "--") {
		err := indexFileName(fsPath, fsName)
		if err != nil {
			return true, http.StatusInternalServerError, fmt.Errorf("failed indexing: %v", err)
		}
	}
	return false, 0, nil
}

func indexPlaintext(cascade bool, fsPath string, fsName string, originalFsPath string, originalFsName string) (bool, HttpError, error) {
	fullName := fsPath + fsName
	if IsTextFile(fullName) && cascade {

		f, err := fs.F.Open(fullName)
		if err != nil {
			return true, http.StatusInternalServerError, fmt.Errorf("Could not open file for indexing %s: %v", fullName, err)
		}
		defer f.Close()

		// parts 1 and up index the binary data
		var rdr io.Reader = f
		buffer := make([]byte, 16*1024)
		part := 1
		for {
			sz, err := rdr.Read(buffer)
			if err == io.EOF {
				break
			}
			err = indexTextFile(fsPath, fsName, part, originalFsPath, originalFsName, buffer[:sz])
			if err != nil {
				return true, http.StatusInternalServerError, fmt.Errorf("failed indexing: %v", err)
			}
			part++
		}
		return true, http.StatusOK, nil
	}
	return false, 0, nil
}

func extractImage(cascade bool, fsName string, user data.User, fsPath string, originalFsPath string, originalFsName string, privileged bool) (bool, HttpError, error) {
	fullName := fsPath + fsName
	if IsImage(fullName) && cascade {
		if true {
			rdr, err := fs.F.MakeThumbnail(fullName)
			if err != nil {
				return true, http.StatusInternalServerError, fmt.Errorf("Could not make thumbnail for %s: %v", fullName, err)
			}
			thumbnailName := fmt.Sprintf("%s--thumbnail.png", fsName)
			herr, err := postFileHandler(user, rdr, fsPath, thumbnailName, originalFsPath, originalFsName, false, privileged)
			if err != nil {
				return true, herr, fmt.Errorf("Could not write make thumbnail for indexing %s: %v", fullName, err)
			}
		}

		if os.Getenv("AWS_ACCESS_KEY_ID") != "" {
			{
				log.Printf("detect labels on %s", fullName)
				rdr, err := detectLabels(fullName)
				if err != nil {
					return true, http.StatusInternalServerError, fmt.Errorf("Could not extract labels for %s: %v", fullName, err)
				}
				labelName := fmt.Sprintf("%s--labels.json", fsName)
				herr, err := postFileHandler(user, rdr, fsPath, labelName, originalFsPath, originalFsName, cascade, privileged)
				if err != nil {
					return true, herr, fmt.Errorf("Could not write labgel detect %s: %v", fullName, err)
				}

				labelFile := fullName + "--labels.json"
				jf, err := fs.F.ReadFile(labelFile)
				if err != nil {
					return true, http.StatusInternalServerError, fmt.Errorf("Could not find file: %s %v", labelFile, err)
				}
				var j LabelModel
				err = json.Unmarshal(jf, &j)
				if err != nil {
					return true, http.StatusInternalServerError, fmt.Errorf("Could not look for celeb detect on labels for %s", fullName)
				} else {
					for i := range j.Labels {
						v := j.Labels[i].Name
						if v == "Face" || v == "Person" || v == "People" {
							log.Printf("detect celebs on %s", fullName)
							rdr, err = detectCeleb(fullName)
							if err != nil {
								return true, http.StatusInternalServerError, fmt.Errorf("Could not extract labels for %s: %v", fullName, err)
							}
							if rdr != nil {
								faceName := fmt.Sprintf("%s--celebs.json", fsName)
								herr, err := postFileHandler(user, rdr, fsPath, faceName, originalFsPath, originalFsName, cascade, privileged)
								if err != nil {
									return true, herr, fmt.Errorf("Could not write face detect %s: %v", fullName, err)
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
					return true, http.StatusInternalServerError, fmt.Errorf("Could not do content moderation for %s: %v", fullName, err)
				}
				labelName := fmt.Sprintf("%s--moderation.json", fsName)
				herr, err := postFileHandler(user, rdr, fsPath, labelName, originalFsPath, originalFsName, cascade, privileged)
				if err != nil {
					return true, herr, fmt.Errorf("Could not content moderate %s: %v", fullName, err)
				}
			}
		}

		return true, http.StatusOK, nil
	}
	return false, 0, nil
}

func extractVideo(cascade bool, fsName string, user data.User, fsPath string, originalFsPath string, originalFsName string, privileged bool) (bool, HttpError, error) {
	fullName := fsPath + fsName
	if IsVideo(fullName) && cascade {
		rdr, err := fs.F.VideoThumbnail(fullName)
		if err != nil {
			return true, http.StatusInternalServerError, fmt.Errorf("Could not make thumbnail for %s: %v", fullName, err)
		}
		thumbnailName := fmt.Sprintf("%s--thumbnail.png", fsName)
		herr, err := postFileHandler(user, rdr, fsPath, thumbnailName, originalFsPath, originalFsName, false, privileged)
		if err != nil {
			return true, herr, fmt.Errorf("Could not write make thumbnail for indexing %s: %v", fullName, err)
		}
		return true, http.StatusOK, nil
	}
	return false, 0, nil
}

func extractDoc(cascade bool, fsName string, user data.User, fsPath string, originalFsPath string, originalFsName string, privileged bool) (bool, HttpError, error) {
	fullName := fsPath + fsName
	if IsDoc(fullName) && cascade {

		f, err := fs.F.Open(fullName)
		if err != nil {
			return true, http.StatusInternalServerError, fmt.Errorf("Could not open file for indexing %s: %v", fullName, err)
		}

		rdr, err := DocExtract(fullName, f)
		f.Close()
		if err != nil {
			return true, http.StatusInternalServerError, fmt.Errorf("Could not extract file for indexing %s: %v", fullName, err)
		}

		extractName := fmt.Sprintf("%s--extract.txt", fsName)
		herr, err := postFileHandler(user, rdr, fsPath, extractName, originalFsPath, originalFsName, cascade, privileged)
		if err != nil {
			return true, herr, fmt.Errorf("Could not write extract file for indexing %s: %v", fullName, err)
		}

		ext := strings.ToLower(path.Ext(fullName))
		if ext == ".pdf" {
			rdr, err := fs.F.PdfThumbnail(fullName)
			if err != nil {
				return true, http.StatusInternalServerError, fmt.Errorf("Could not make thumbnail for %s: %v", fullName, err)
			}

			thumbnailName := fmt.Sprintf("%s--thumbnail.png", fsName)
			herr, err := postFileHandler(user, rdr, fsPath, thumbnailName, originalFsPath, originalFsName, false, privileged)
			if err != nil {
				return true, herr, fmt.Errorf("Could not write make thumbnail for indexing %s: %v", fullName, err)
			}
		}

		return true, http.StatusOK, nil
	}
	return false, 0, nil
}

func putFileOnDisk(fullName string, stream io.Reader) (bool, HttpError, error) {
	if true {
		f, err := fs.F.Create(fullName)
		if err != nil {
			return true, http.StatusInternalServerError, fmt.Errorf("Could not create file %s: %v", fullName, err)
		}

		sz, err := io.Copy(f, stream)
		f.Close()
		if err != nil {
			return true, http.StatusInternalServerError, fmt.Errorf("Could not write to file (%d bytes written) %s: %v", sz, fullName, err)
		}

		recompileTemplates(fullName)
	}
	return false, 0, nil
}

func badUploadPermissions(fsName string, stream io.Reader, user data.User) (io.Reader, bool, HttpError, error) {
	if fsName == "permissions.rego" || strings.HasSuffix(fsName, "--permissions.rego") {
		proposedUpload, err := ioutil.ReadAll(stream)
		if err != nil {
			return nil, true, http.StatusForbidden, fmt.Errorf(
				"Could not read proposed permissions.rego: %v",
				err,
			)
		}
		proposedAttrs, err := CalculateRego(user, string(proposedUpload))
		if err != nil {
			return nil, true, http.StatusForbidden, fmt.Errorf(
				"Could not calculate proposed permissions.rego: %v",
				err,
			)
		}
		if !proposedAttrs.Write || !proposedAttrs.Read {
			return nil, true, http.StatusForbidden, fmt.Errorf(
				"Proposed permissions.rego does not allow write and read: %v",
				proposedAttrs,
			)
		}
		stream = bytes.NewReader(proposedUpload)
	}
	return stream, false, 0, nil
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
