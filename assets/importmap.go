package assets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path"
	"strings"
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
		Imports map[string]string            `json:"imports"`
		Scopes  map[string]map[string]string `json:"scopes,omitempty"`
	}

	if err := json.NewDecoder(f).Decode(&importMap); err != nil {
		return "", err
	}

	// Process imports paths through fingerprinting
	for k, v := range importMap.Imports {
		hashed, err := m.PathFor(v)
		if err != nil {
			fmt.Printf("[error] error resolving %q: %v\n", v, err)
			continue
		}

		importMap.Imports[k] = hashed
	}

	// Process scopes: transform scope keys to include serving path and fingerprint values
	if len(importMap.Scopes) > 0 {
		transformedScopes := make(map[string]map[string]string)
		
		for scopeKey, scopeMap := range importMap.Scopes {
			// Transform the scope key to include serving path prefix
			// e.g., "vendor/@org/package@1.2.3/" -> "/assets/vendor/@org/package@1.2.3/"
			transformedKey := m.addServingPath(scopeKey)
			
			transformedScopes[transformedKey] = make(map[string]string)
			
			// Process each scope value through fingerprinting
			for k, v := range scopeMap {
				hashed, err := m.PathFor(v)
				if err != nil {
					fmt.Printf("[error] error resolving scope path %q: %v\n", v, err)
					continue
				}

				transformedScopes[transformedKey][k] = hashed
			}
		}
		
		importMap.Scopes = transformedScopes
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

// addServingPath adds the serving path prefix to a scope key
// e.g., "vendor/@org/package@1.2.3/" -> "/assets/vendor/@org/package@1.2.3/"
func (m *manager) addServingPath(scopeKey string) string {
	// Clean up the scope key
	scopeKey = strings.TrimPrefix(scopeKey, "/")
	scopeKey = strings.TrimSuffix(scopeKey, "/")
	
	// Add serving path prefix and trailing slash
	return path.Join("/", m.servingPath, scopeKey) + "/"
}
