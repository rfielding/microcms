package handler

import (
	"io"

	"github.com/rfielding/microcms/fs"
)

/*
These assume real volume mount. And it may be fixable through temp files.
*/

func MakeThumbnail(file string) (io.Reader, error) {
	command := []string{
		"convert",
		"-thumbnail", "x100",
		"-background", "white",
		"-alpha", "remove",
		"-format", "png",
		(fs.F.At() + file),
		"-",
	}
	return CommandReader(file, command)
}

func VideoThumbnail(file string) (io.Reader, error) {
	command := []string{
		"convert",
		"-resize", "x100",
		(fs.F.At() + file + "[100]"),
		"png:-",
	}
	return CommandReader(file, command)
}
