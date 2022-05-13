package main

import (
	"context"
	"io"
	"os"

	vision "cloud.google.com/go/vision/apiv1"
)

// detectLabels gets labels from the Vision API for an image at the given file path.
func detectLabels(file string) (io.Reader, error) {
	ctx := context.Background()

	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	image, err := vision.NewImageFromReader(f)
	if err != nil {
		return nil, err
	}
	annotations, err := client.DetectLabels(ctx, image, nil, 10)
	if err != nil {
		return nil, err
	}

	pipeReader, pipeWriter := io.Pipe()
	go func() {
		pipeWriter.Write([]byte(AsJson(annotations)))
		pipeWriter.Close()
	}()
	return pipeReader, nil
}
