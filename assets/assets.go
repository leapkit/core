package assets

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed public.tmpl
var tmpl []byte

func generateEmbed(publicFolder string) error {
	var embs []string
	err := filepath.Walk(publicFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(publicFolder, path)
		if err != nil {
			return err
		}

		embs = append(embs, fmt.Sprintf("`%v`", filepath.ToSlash(relativePath)))
		return nil
	})

	if err != nil {
		return err
	}

	f, err := os.Create("./public/public.go")
	if err != nil {
		return err
	}

	files := strings.Join(embs, " ")
	tt := template.Must(template.New("public").Parse(string(tmpl)))
	err = tt.Execute(f, files)

	return err
}
