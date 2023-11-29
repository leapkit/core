package assets

import (
	"os"
)

// Embed the assets in the destination folder, this function
// first copy the files and then generates a public.go file
// with the assets embedded through a go:embed directive.
func Embed(src, dst string) error {
	// Remove the destination folder before copying the files
	err := os.RemoveAll(dst)
	if err != nil {
		return err
	}

	// Copy and generate the files
	err = copyFiles(src, dst)
	if err != nil {
		panic(err)
	}

	//Generate public.go
	err = generateEmbed(dst)
	if err != nil {
		return err
	}

	return nil
}
