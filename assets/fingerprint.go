package assets

import (
	"crypto/md5"
	"encoding/hex"
	"path"
)

// PathFor returns the fingerprinted path for a given
// file.
func (m *manager) PathFor(fname string) string {
	if m.fileToHash[fname] != "" {
		return path.Join(path.Dir(fname), m.fileToHash[fname])
	}

	// Compute the hash of the file
	bb, err := m.ReadFile(fname)
	if err != nil {
		return fname
	}

	hash := md5.Sum(bb)
	hashString := hex.EncodeToString(hash[:])

	// Add the hash to the filename
	filename := path.Base(fname)
	ext := path.Ext(fname)
	newFilename := filename[:len(filename)-len(ext)] + "-" + hashString + ext

	m.fmut.Lock()
	defer m.fmut.Unlock()
	m.fileToHash[fname] = newFilename
	m.HashToFile[newFilename] = fname

	return path.Join(path.Dir(fname), newFilename)
}
