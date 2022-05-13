package main

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

// postFileHandler can be re-used as long as err != nil
func postFileHandler(
	w http.ResponseWriter,
	r *http.Request,
	stream io.Reader,
	command string,
	parentDir string,
	name string,
	originalParentDir string,
	originalName string,
	cascade bool,
) error {
	fullName := fmt.Sprintf("%s/%s", parentDir, name)
	//log.Printf("create %s %s", command, fullName)

	// Write the file out
	flags := os.O_WRONLY | os.O_CREATE

	// We either append content, or overwrite it entirely
	if command == "append" {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	// TODO: check permissions before allowing writes

	//log.Printf("Ensure existence of parentDir: %s", parentDir)
	err := os.MkdirAll("."+parentDir, 0777)
	if err != nil {
		return HandleReturnedError(w, err, "Could not create path for %s: %v", r.URL.Path)
	}

	existingSize := int64(0)
	s, err := os.Stat("." + fullName)
	if err == nil {
		existingSize = s.Size()
	}

	// Ensure that the file in question exists on disk.
	if true {
		f, err := os.Create("." + fullName) //, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return HandleReturnedError(w, err, "Could not create file %s: %v", r.URL.Path)
		}

		// Save the stream to a file
		sz, err := io.Copy(f, stream)
		f.Close() // strange positioning, but we must close before defer can get to it.
		if err != nil {
			return HandleReturnedError(w, err, "Could not write to file (%d bytes written) %s: %v", sz, r.URL.Path)
		}
	}

	if IsDoc(fullName) && cascade {
		// Open the file we wrote
		f, err := os.Open("." + fullName)
		if err != nil {
			return HandleReturnedError(w, err, "Could not open file for indexing %s: %v", fullName)
		}
		// Get a doc extract stream
		rdr, err := DocExtract(fullName, f)
		f.Close()
		if err != nil {
			return HandleReturnedError(w, err, "Could not extract file for indexing %s: %v", fullName)
		}
		// Write the doc extract stream like an upload
		extractName := fmt.Sprintf("%s--extract.txt", name)
		err = postFileHandler(w, r, rdr, command, parentDir, extractName, originalParentDir, originalName, cascade)
		if err != nil {
			return HandleReturnedError(w, err, "Could not write extract file for indexing %s: %v", fullName)
		}

		ext := strings.ToLower(path.Ext(fullName))
		if ext == ".pdf" {
			rdr, err := pdfThumbnail(`./` + fullName)
			if err != nil {
				return HandleReturnedError(w, err, "Could not make thumbnail for %s: %v", fullName)
			}
			// Only png works.  bug in imageMagick.  don't cascade on thumbnails
			thumbnailName := fmt.Sprintf("%s--thumbnail.png", name)
			err = postFileHandler(w, r, rdr, command, parentDir, thumbnailName, originalParentDir, originalName, false)
			if err != nil {
				return HandleReturnedError(w, err, "Could not write make thumbnail for indexing %s: %v", fullName)
			}
		}

		// open the file that we saved, and index it in the database.
		return nil
	}

	if IsVideo(fullName) && cascade {
		rdr, err := videoThumbnail(`./` + fullName)
		if err != nil {
			return HandleReturnedError(w, err, "Could not make thumbnail for %s: %v", fullName)
		}
		thumbnailName := fmt.Sprintf("%s--thumbnail.png", name)
		err = postFileHandler(w, r, rdr, command, parentDir, thumbnailName, originalParentDir, originalName, false)
		if err != nil {
			return HandleReturnedError(w, err, "Could not write make thumbnail for indexing %s: %v", fullName)
		}
		return nil
	}

	if IsImage(fullName) && cascade {
		if true {
			rdr, err := makeThumbnail(`./` + fullName)
			if err != nil {
				return HandleReturnedError(w, err, "Could not make thumbnail for %s: %v", fullName)
			}
			thumbnailName := fmt.Sprintf("%s--thumbnail.png", name)
			err = postFileHandler(w, r, rdr, command, parentDir, thumbnailName, originalParentDir, originalName, false)
			if err != nil {
				return HandleReturnedError(w, err, "Could not write make thumbnail for indexing %s: %v", fullName)
			}
		}

		if useVisionAPI {
			rdr, err := detectLabels(`./` + fullName)
			if err != nil {
				return HandleReturnedError(w, err, "Could not extract labels for %s: %v", fullName)
			}
			labelName := fmt.Sprintf("%s--labels.json", name)
			err = postFileHandler(w, r, rdr, command, parentDir, labelName, originalParentDir, originalName, cascade)
			if err != nil {
				//return HandleReturnedError(w, err, "Could not write extract file for indexing %s: %v", fullName)
				log.Printf("Could not write extract file for indexing %s: %v\n", fullName, err)
				return nil
			}
		}

		return nil
	}

	if IsTextFile(fullName) && cascade {
		// open the file that we saved, and index it in the database.
		f, err := os.Open("." + fullName)
		if err != nil {
			return HandleReturnedError(w, err, "Could not open file for indexing %s: %v", fullName)
		}
		defer f.Close()
		if existingSize > 0 {
			// we are appending, so we need to start at the end of the file
			f.Seek(existingSize, 0)
		}
		var rdr io.Reader = f
		/*
			if command == "files" {
				// this implies a truncate
				_, err := theDB.Exec(`DELETE from filesearch where path = ? and name = ? and cmd = ?`, parentDir+"/", name, command)
				if err != nil {
					log.Printf("cleaning out fulltextsearch for: %s%s %s failed: %v", parentDir+"/", name, command, err)
				}
			}
		*/
		buffer := make([]byte, 4*1024)
		part := 0
		for {
			sz, err := rdr.Read(buffer)
			if err == io.EOF {
				break
			}
			err = indexTextFile(command, parentDir+"/", name, part, originalParentDir+"/", originalName, buffer[:sz])
			if err != nil {
				log.Printf("failed indexing: %v", err)
			}
			part++
		}
		return nil
	}
	return nil
}

func postFilesHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) {
	var err error
	defer r.Body.Close()

	q := r.URL.Query()
	// This is a signal that this is a tar archive
	// that we unpack to install all files at the given url
	needsInstall := q.Get("install") == "true"
	if needsInstall {
		log.Printf("install tarball to %s", r.URL.Path)
	}

	if len(pathTokens) < 2 {
		HandleError(w, err, "path needs /[command]/[url] for post to %s: %v", r.URL.Path)
		return
	}
	command := pathTokens[1]
	pathTokens[1] = "files"

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
				err = postFileHandler(w, r, t, command, tardir, tarname, tardir, tarname, true)
				if err != nil {
					log.Printf("ERR %v", err)
					return
				}
			}
		}
	} else {
		// Just a normal single-file upload
		err = postFileHandler(w, r, r.Body, command, parentDir, name, parentDir, name, true)
		if err != nil {
			log.Printf("ERR %v", err)
			return
		}
	}
}
