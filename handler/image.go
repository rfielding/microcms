package handler

import (
	"io"

	"github.com/rfielding/microcms/fs"
)

/*
These assume real volume mount. And it may be fixable through temp files.
*/

func MakeThumbnail(fullName string) (io.Reader, error) {
	command := []string{
		"convert",
		"-thumbnail", "x100",
		"-background", "white",
		"-alpha", "remove",
		"-format", "png",
		(fs.F.At() + fullName),
		"-",
	}
	return CommandReader(fullName, command)
}

func VideoThumbnail(fullName string) (io.Reader, error) {
	command := []string{
		"convert",
		"-resize", "x100",
		(fs.F.At() + fullName + "[100]"),
		"png:-",
	}
	return CommandReader(fullName, command)
}
