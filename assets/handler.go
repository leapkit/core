package assets

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func (m *manager) HandlerPattern() string {
	if m.servingPath == "/" {
		return ""
	}

	return m.servingPath
}

// Open opens the named file. This method is part of the fs.FS interface implementation.
// It handles special cases for assets, such as preventing access to Go source files and
// resolving hashed filenames to their original forms.
//
// This function implements the fs.FS interface and allows using the manager as a filesystem
// compatible with standard library functions that expect an fs.FS.
func (m *manager) Open(name string) (fs.File, error) {
	ext := filepath.Ext(name)
	if ext == ".go" {
		return nil, os.ErrNotExist
	}

	// Converting hashed into original file name
	if original, ok := m.HashToFile[name]; ok {
		name = original
	}

	name = strings.TrimPrefix(name, m.servingPath)

	return m.embedded.Open(name)
}

func (m *manager) ReadFile(name string) ([]byte, error) {
	x, err := m.Open(name)
	if err != nil {
		return nil, err
	}
	defer x.Close()

	return io.ReadAll(x)
}
