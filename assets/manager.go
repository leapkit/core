package assets

import (
	"io/fs"
	"path"
	"strings"
	"sync"
)

type manager struct {
	embedded fs.FS

	servingPath string

	fmut       sync.Mutex
	fileToHash map[string]string
	HashToFile map[string]string
}

// NewManager returns a new manager that wraps the given fs.FS.
func NewManager(embedded fs.FS, servingPath string) *manager {
	servingPath = path.Clean(servingPath)
	if servingPath == "." {
		servingPath = "/"
	}

	servingPath = strings.Trim(servingPath, "/")
	if servingPath == "" {
		servingPath = "/"
	} else {
		servingPath = "/" + servingPath + "/"
	}

	return &manager{
		embedded:    embedded,
		servingPath: servingPath,
		fileToHash:  map[string]string{},
		HashToFile:  map[string]string{},
	}
}
