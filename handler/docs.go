package handler

import (
	"fmt"
	"io"
	"net/http"
)

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
