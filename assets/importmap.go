package assets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
)

func (m *manager) ImportMap() (template.HTML, error) {
	f, err := m.Open("importmap.json")
	if os.IsNotExist(err) {
		return "", nil
	}

	if err != nil {
		return "", err
	}

	defer f.Close()

	var importMap struct {
		Imports map[string]string `json:"imports"`
	}

	if err := json.NewDecoder(f).Decode(&importMap); err != nil {
		return "", err
	}

	for k, v := range importMap.Imports {
		hashed, err := m.PathFor(v)
		if err != nil {
			fmt.Printf("[error] error resolving %q: %v\n", v, err)
			continue
		}

		importMap.Imports[k] = hashed
	}

	b, err := json.MarshalIndent(importMap, "", "  ")
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(nil)
	buf.WriteString(`<script type="importmap">`)
	buf.WriteString("\n")
	buf.Write(b)
	buf.WriteString("</script>")

	if _, ok := importMap.Imports["application"]; ok {
		buf.WriteString("\n")
		buf.WriteString(`<script type="module">import "application";</script>`)
	}

	return template.HTML(buf.String()), nil
}
