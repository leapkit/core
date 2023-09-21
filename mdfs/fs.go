// package mdfs provides a file system that wraps a directory and an embed.FS.
// it prioritizes the directory FS and falls back to the embedded FS if the file cannot
// be found on disk. This is useful during development or when deploying with
// assets not embedded in the binary.
package mdfs

import (
	"io"
	"io/fs"
	"os"
)

// FS wraps a directory and an embed FS that are expected to have the same contents.
// it prioritizes the directory FS and falls back to the embedded FS if the file cannot
// be found on disk. This is useful during development or when deploying with
// assets not embedded in the binary.
type FS struct {
	dir string

	embed fs.FS
	dirFs fs.FS

	useLocal bool
}

// NewFallbackFS returns a new FS that wraps the given directory and embedded FS.
// the embed.FS is expected to embed the same files as the directory FS.
func New(embed fs.FS, dir, env string) FS {
	// If the directory is empty, use the current working directory.
	if dir == "" {
		pwd, _ := os.Getwd()
		dir = pwd
	}

	return FS{
		embed: embed,
		dirFs: os.DirFS(dir),
		dir:   dir,

		useLocal: (env == "development"),
	}
}

// Open opens the named file.
//
// When Open returns an error, it should be of type *PathError with the Op
// field set to "open", the Path field set to name, and the Err field
// describing the problem.
//
// Open should reject attempts to open names that do not satisfy
// ValidPath(name), returning a *PathError with Err set to ErrInvalid or
// ErrNotExist.
func (f FS) Open(name string) (file fs.File, err error) {
	if f.useLocal {
		file, err = f.dirFs.Open(name)
		if err == nil {
			return
		}
	}

	file, err = f.embed.Open(name)
	return file, err
}

// ReadFile reads the named file from the file system fs and returns its contents.
// It uses the custom Open method to open the file.
func (f FS) ReadFile(name string) ([]byte, error) {
	x, err := f.Open(name)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(x)
}
