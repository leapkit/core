package assets

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func copyFiles(source, destination string) error {
	var exceptions = []string{
		// Todo: configurable exceptions
		"assets/application.css",
	}

	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Get the relative path of the file
		relativePath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		for _, v := range exceptions {
			if strings.HasSuffix(relativePath, v) {
				return nil
			}
		}

		// Create the destination folder if it doesn't exist
		destFolder := filepath.Join(destination, filepath.Dir(relativePath))
		err = os.MkdirAll(destFolder, os.ModePerm)
		if err != nil {
			return err
		}

		// Copy the file to the destination folder
		destPath := filepath.Join(destFolder, filepath.Base(relativePath))
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, srcFile)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
