package assets_test

import (
	"html/template"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"go.leapkit.dev/core/assets"
)

func TestFingerprint(t *testing.T) {
	assetsPath := "/files"
	m := assets.NewManager(fstest.MapFS{
		"main.js":        {Data: []byte("AAA")},
		"other/main.js":  {Data: []byte("AAA")},
		"custom/main.go": {Data: []byte("AAAA")},
	}, assetsPath)

	t.Run("is deterministic", func(t *testing.T) {
		a, _ := m.PathFor("files/main.js")
		b, _ := m.PathFor("files/main.js")
		if a != b {
			t.Errorf("Expected %s to equal %s", a, b)
		}

		if !strings.Contains(a, assetsPath) {
			t.Errorf("Expected %s to have %s prefix", a, assetsPath)
		}

		a, _ = m.PathFor("files/other/main.js")
		b, _ = m.PathFor("files/other/main.js")
		if a != b {
			t.Errorf("Expected %s to equal %s", a, b)
		}

		if !strings.Contains(a, assetsPath) {
			t.Errorf("Expected %s to have %s prefix", a, assetsPath)
		}
	})

	t.Run("concurrent PathFor access", func(t *testing.T) {
		const goroutines = 20
		const iterations = 100
		m := assets.NewManager(fstest.MapFS{
			"main.js":        {Data: []byte("AAA")},
			"other/main.js":  {Data: []byte("BBB")},
			"custom/main.go": {Data: []byte("CCCC")},
		}, "/files")

		files := []string{
			"main.js",
			"other/main.js",
			"custom/main.go",
			"/files/main.js",
			"/files/other/main.js",
			"/files/custom/main.go",
		}

		ch := make(chan struct{}, goroutines)
		for i := range goroutines {
			go func(id int) {
				defer func() {
					ch <- struct{}{}
				}()

				for j := 0; j < iterations; j++ {
					file := files[(id+j)%len(files)]
					_, _ = m.PathFor(file)
				}
			}(i)
		}

		for range goroutines {
			<-ch
		}
	})

	t.Run("adds starting slash", func(t *testing.T) {
		a, err := m.PathFor("files/main.js")
		if err != nil {
			t.Fatal(err)
		}

		b, err := m.PathFor("/files/main.js")
		if err != nil {
			t.Fatal(err)
		}

		if a != b {
			t.Errorf("Expected %s to equal %s", a, b)
		}
	})

	t.Run("adds starting /files", func(t *testing.T) {
		a, _ := m.PathFor("main.js")
		t.Log(a)
		if !strings.HasPrefix(a, "/files") {
			t.Errorf("Expected %s to start with /files", a)
		}
	})

	t.Run("respects folders", func(t *testing.T) {
		a, err := m.PathFor("main.js")
		if err != nil {
			t.Fatal(err)
		}

		if !strings.HasPrefix(a, "/files/main-") {
			t.Errorf("Expected %s to contain /files/other/main-<hash>", a)
		}

		b, _ := m.PathFor("files/other/main.js")
		if err != nil {
			t.Fatal(err)
		}

		if !strings.HasPrefix(b, "/files/other/main-") {
			t.Errorf("Expected %s to contain /files/other/main-<hash>", b)
		}

		if a == b {
			t.Errorf("Expected %s to not equal %s", a, b)
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		a, err := m.PathFor("foo.js")
		if err == nil {
			t.Errorf("File must not exists: %s", a)
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		a, err := m.PathFor("custom/main.go")
		if err == nil {
			t.Errorf("File must not exists: %s", a)
		}
	})

	t.Run("manager looks for files in root", func(t *testing.T) {
		m := assets.NewManager(fstest.MapFS{
			"main.js":       {Data: []byte("AAA")},
			"other/main.js": {Data: []byte("AAA")},
		}, "")

		a, err := m.PathFor("main.js")
		if err != nil {
			t.Fatal(err)
		}

		if !strings.HasPrefix(a, "/") {
			t.Errorf("Expected %s to start with /", a)
		}

		b, err := m.PathFor("other/main.js")
		if err != nil {
			t.Fatal(err)
		}

		if !strings.HasPrefix(b, "/") {
			t.Errorf("Expected %s to start with /", b)
		}
	})

	t.Run("HandlerPattern", func(t *testing.T) {
		cases := []struct {
			pattern  string
			expected string
		}{
			{"/", ""},
			{"/", ""},

			{"files", "/files/"},
			{"/files/", "/files/"},
			{"files/", "/files/"},
			{"/files/", "/files/"},

			{"files/path", "/files/path/"},
			{"/files/path", "/files/path/"},
			{"files/path/", "/files/path/"},
			{"/files/path/", "/files/path/"},
		}

		for _, c := range cases {
			m := assets.NewManager(fstest.MapFS{}, c.pattern)
			if m.HandlerPattern() != c.expected {
				t.Errorf("Expected %q to equal %q", m.HandlerPattern(), c.expected)
			}
		}
	})

	t.Run("HandlerPattern with root serving path", func(t *testing.T) {
		m := assets.NewManager(fstest.MapFS{}, "/")
		if m.HandlerPattern() == "/" {
			t.Errorf("Expected empty string, got %s", m.HandlerPattern())
		}

		m = assets.NewManager(fstest.MapFS{}, "public")
		if m.HandlerPattern() != "/public/" {
			t.Errorf("Expected '/public/', got %s", m.HandlerPattern())
		}
	})

	t.Run("PathFor normalize file path", func(t *testing.T) {
		cases := []struct {
			servingPath string
			name        string
			expected    string
		}{
			{"public", "main.js", "/public/main-"},
			{"public", "/main.js", "/public/main-"},
			{"public", "public/main.js", "/public/main-"},
			{"public", "/public/main.js", "/public/main-"},
			{"public", "/other/main.js", "/public/other/main-"},
			{"public", "other/main.js", "/public/other/main-"},

			{"/public", "main.js", "/public/main-"},
			{"/public", "/main.js", "/public/main-"},
			{"/public", "public/main.js", "/public/main-"},
			{"/public", "/public/main.js", "/public/main-"},
			{"/public", "public/other/main.js", "/public/other/main-"},
			{"/public", "/public/other/main.js", "/public/other/main-"},

			{"/public/", "main.js", "/public/main-"},
			{"/public/", "/main.js", "/public/main-"},
			{"/public/", "public/main.js", "/public/main-"},
			{"/public/", "/public/main.js", "/public/main-"},
			{"/public/", "public/other/main.js", "/public/other/main-"},
			{"/public/", "/public/other/main.js", "/public/other/main-"},

			{"public/other", "main.js", "/public/other/main-"},
			{"public/other", "/main.js", "/public/other/main-"},
			{"public/other", "public/other/main.js", "/public/other/main-"},
			{"public/other", "/public/other/main.js", "/public/other/main-"},
			{"/public/other", "public/other/main.js", "/public/other/main-"},
			{"/public/other", "/public/other/main.js", "/public/other/main-"},

			{"/public/other", "main.js", "/public/other/main-"},
			{"/public/other", "/main.js", "/public/other/main-"},
			{"/public/other", "public/other/main.js", "/public/other/main-"},
			{"/public/other", "/public/other/main.js", "/public/other/main-"},

			{"/public/other/", "main.js", "/public/other/main-"},
			{"/public/other/", "/main.js", "/public/other/main-"},
			{"/public/other/", "public/other/main.js", "/public/other/main-"},
			{"/public/other/", "/public/other/main.js", "/public/other/main-"},
		}

		for i, c := range cases {
			manager := assets.NewManager(fstest.MapFS{
				"main.js":       {Data: []byte("AAA")},
				"other/main.js": {Data: []byte("AAA")},
			}, c.servingPath)

			result, err := manager.PathFor(c.name)
			if err != nil {
				t.Errorf("%d, Expected no error, got %s", i, err)
			}

			if !strings.HasPrefix(result, c.expected) {
				t.Errorf("%d, Expected %s to start with /public/", i, result)
			}
		}
	})

	t.Run("recalculate hashed files only un development", func(t *testing.T) {
		fsMap := fstest.MapFS{
			"main.js": {Data: []byte("AAA")},
		}

		m := assets.NewManager(fsMap, assetsPath)

		t.Run("in development mode should recalculate the asset hash", func(t *testing.T) {
			t.Setenv("GO_ENV", "development")

			a, err := m.PathFor("main.js")
			if err != nil {
				t.Errorf("Expected nil, got %s", err)
			}

			fsMap["main.js"] = &fstest.MapFile{Data: []byte("BBB")}

			b, err := m.PathFor("main.js")
			if err != nil {
				t.Errorf("Expected nil, got %s", err)
			}

			if a == b {
				t.Errorf("Expected %s to not equal %s", a, b)
			}
		})

		t.Run("in other env should not recalculate the asset hash", func(t *testing.T) {
			t.Setenv("GO_ENV", "production")

			a, err := m.PathFor("main.js")
			if err != nil {
				t.Errorf("Expected nil, got %s", err)
			}

			fsMap["main.js"] = &fstest.MapFile{Data: []byte("CCC")}

			b, err := m.PathFor("main.js")
			if err != nil {
				t.Errorf("Expected nil, got %s", err)
			}

			if a != b {
				t.Errorf("Expected %s to equal %s", a, b)
			}
		})
	})

	t.Run("Path", func(t *testing.T) {
		t.Run("returns fingerprinted path on success", func(t *testing.T) {
			fingerprintedPath, err := m.PathFor("main.js")
			if err != nil {
				t.Fatalf("PathFor failed: %v", err)
			}

			path := m.Path("main.js")
			if path != fingerprintedPath {
				t.Errorf("expected %q, got %q", fingerprintedPath, path)
			}
		})

		t.Run("returns original name on error", func(t *testing.T) {
			originalName := "nonexistent.js"
			path := m.Path(originalName)

			if path != originalName {
				t.Errorf("expected %q, got %q", originalName, path)
			}

			// Also check the error from PathFor to be sure
			_, err := m.PathFor(originalName)
			if err == nil {
				t.Errorf("expected PathFor to return an error for %q", originalName)
			}
		})
	})
}

func TestReadFile(t *testing.T) {
	t.Run("successfully reads file", func(t *testing.T) {
		content := "console.log('hello world');"
		m := assets.NewManager(fstest.MapFS{
			"main.js": {Data: []byte(content)},
		}, "/assets")

		data, err := m.ReadFile("main.js")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if string(data) != content {
			t.Errorf("Expected %q, got %q", content, string(data))
		}
	})

	t.Run("file not found", func(t *testing.T) {
		m := assets.NewManager(fstest.MapFS{
			"main.js": {Data: []byte("content")},
		}, "/assets")

		_, err := m.ReadFile("nonexistent.js")
		if err == nil {
			t.Error("Expected error when reading nonexistent file")
		}
	})

	t.Run("cannot read Go files", func(t *testing.T) {
		m := assets.NewManager(fstest.MapFS{
			"main.go": {Data: []byte("package main")},
		}, "/assets")

		_, err := m.ReadFile("main.go")
		if err == nil {
			t.Error("Expected error when reading .go file")
		}
	})

	t.Run("reads file with serving path prefix", func(t *testing.T) {
		content := "test content"
		m := assets.NewManager(fstest.MapFS{
			"test.js": {Data: []byte(content)},
		}, "/assets")

		data, err := m.ReadFile("/assets/test.js")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if string(data) != content {
			t.Errorf("Expected %q, got %q", content, string(data))
		}
	})
}

// errorFS is a test filesystem that returns a custom error for importmap.json
type errorFS struct{}

func (e *errorFS) Open(name string) (fs.File, error) {
	if name == "importmap.json" {
		return nil, fs.ErrPermission // Return a non-NotExist error
	}
	return nil, fs.ErrNotExist
}

func TestImportMap(t *testing.T) {
	t.Run("no importmap.json file", func(t *testing.T) {
		m := assets.NewManager(fstest.MapFS{
			"main.js": {Data: []byte("console.log('hello');")},
		}, "/assets")

		result, err := m.ImportMap()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result != "" {
			t.Errorf("Expected empty string, got %q", result)
		}
	})

	t.Run("valid importmap.json without application", func(t *testing.T) {
		importMapJSON := `{
			"imports": {
				"react": "/assets/react.js",
				"vue": "/assets/vue.js"
			}
		}`

		m := assets.NewManager(fstest.MapFS{
			"importmap.json": {Data: []byte(importMapJSON)},
			"react.js":       {Data: []byte("// React code")},
			"vue.js":         {Data: []byte("// Vue code")},
		}, "/assets")

		result, err := m.ImportMap()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		resultStr := string(result)
		if !strings.Contains(resultStr, `<script type="importmap">`) {
			t.Error("Expected importmap script tag")
		}
		if !strings.Contains(resultStr, "react") {
			t.Error("Expected react import")
		}
		if !strings.Contains(resultStr, "vue") {
			t.Error("Expected vue import")
		}
		if strings.Contains(resultStr, `<script type="module">import "application";</script>`) {
			t.Error("Should not contain application import when not present")
		}
	})

	t.Run("valid importmap.json with application", func(t *testing.T) {
		importMapJSON := `{
			"imports": {
				"application": "/assets/application.js",
				"react": "/assets/react.js"
			}
		}`

		m := assets.NewManager(fstest.MapFS{
			"importmap.json": {Data: []byte(importMapJSON)},
			"application.js": {Data: []byte("// Application code")},
			"react.js":       {Data: []byte("// React code")},
		}, "/assets")

		result, err := m.ImportMap()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		resultStr := string(result)
		if !strings.Contains(resultStr, `<script type="importmap">`) {
			t.Error("Expected importmap script tag")
		}
		if !strings.Contains(resultStr, `<script type="module">import "application";</script>`) {
			t.Error("Expected application import script when application is present")
		}
		if !strings.Contains(resultStr, "application") {
			t.Error("Expected application import")
		}
	})

	t.Run("invalid json in importmap.json", func(t *testing.T) {
		m := assets.NewManager(fstest.MapFS{
			"importmap.json": {Data: []byte("invalid json")},
		}, "/assets")

		_, err := m.ImportMap()
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("asset file not found in imports", func(t *testing.T) {
		importMapJSON := `{
			"imports": {
				"missing": "/assets/missing.js",
				"existing": "/assets/existing.js"
			}
		}`

		m := assets.NewManager(fstest.MapFS{
			"importmap.json": {Data: []byte(importMapJSON)},
			"existing.js":    {Data: []byte("// Existing code")},
		}, "/assets")

		result, err := m.ImportMap()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		resultStr := string(result)
		// Should still generate importmap even if some files are missing
		if !strings.Contains(resultStr, `<script type="importmap">`) {
			t.Error("Expected importmap script tag")
		}
		if !strings.Contains(resultStr, "existing") {
			t.Error("Expected existing import")
		}
	})

	t.Run("returns template.HTML type", func(t *testing.T) {
		importMapJSON := `{
			"imports": {
				"test": "/assets/test.js"
			}
		}`

		m := assets.NewManager(fstest.MapFS{
			"importmap.json": {Data: []byte(importMapJSON)},
			"test.js":        {Data: []byte("// Test code")},
		}, "/assets")

		result, err := m.ImportMap()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify it's actually template.HTML type
		_ = template.HTML(result)
	})

	t.Run("handles file open error other than not exist", func(t *testing.T) {
		// Create a filesystem that will return a different error when opening importmap.json
		m := assets.NewManager(&errorFS{}, "/assets")

		_, err := m.ImportMap()
		if err == nil {
			t.Error("Expected error when file system returns non-NotExist error")
		}
	})

	t.Run("mixed success and failure in PathFor resolution", func(t *testing.T) {
		importMapJSON := `{
			"imports": {
				"existing": "/assets/existing.js",
				"missing1": "/assets/missing1.js",
				"missing2": "/assets/missing2.js"
			}
		}`

		m := assets.NewManager(fstest.MapFS{
			"importmap.json": {Data: []byte(importMapJSON)},
			"existing.js":    {Data: []byte("// Existing code")},
		}, "/assets")

		result, err := m.ImportMap()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		resultStr := string(result)
		if !strings.Contains(resultStr, `<script type="importmap">`) {
			t.Error("Expected importmap script tag")
		}
	})

	t.Run("empty imports map", func(t *testing.T) {
		importMapJSON := `{
			"imports": {}
		}`

		m := assets.NewManager(fstest.MapFS{
			"importmap.json": {Data: []byte(importMapJSON)},
		}, "/assets")

		result, err := m.ImportMap()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		resultStr := string(result)
		if !strings.Contains(resultStr, `<script type="importmap">`) {
			t.Error("Expected importmap script tag")
		}
		if !strings.Contains(resultStr, `"imports": {}`) {
			t.Error("Expected empty imports object")
		}
	})
}

func TestOpen(t *testing.T) {
	t.Run("prevents access to Go files", func(t *testing.T) {
		m := assets.NewManager(fstest.MapFS{
			"main.go":     {Data: []byte("package main")},
			"test.go":     {Data: []byte("package test")},
			"main.js":     {Data: []byte("console.log('hello');")},
			"dir/file.go": {Data: []byte("package dir")},
		}, "/assets")

		// Should not be able to open .go files
		_, err := m.Open("main.go")
		if err == nil {
			t.Error("Expected error when opening .go file")
		}

		_, err = m.Open("test.go")
		if err == nil {
			t.Error("Expected error when opening .go file")
		}

		_, err = m.Open("dir/file.go")
		if err == nil {
			t.Error("Expected error when opening .go file in subdirectory")
		}

		// Should be able to open non-.go files
		_, err = m.Open("main.js")
		if err != nil {
			t.Errorf("Expected no error when opening .js file, got %v", err)
		}
	})

	t.Run("resolves hashed filenames", func(t *testing.T) {
		m := assets.NewManager(fstest.MapFS{
			"main.js": {Data: []byte("console.log('hello');")},
		}, "/assets")

		// First, get the hashed path
		hashedPath, err := m.PathFor("main.js")
		if err != nil {
			t.Fatalf("Failed to get hashed path: %v", err)
		}

		// Extract just the filename from the hashed path
		parts := strings.Split(hashedPath, "/")
		hashedFilename := parts[len(parts)-1]

		// Should be able to open using the hashed filename
		file, err := m.Open(hashedFilename)
		if err != nil {
			t.Errorf("Expected no error when opening hashed filename, got %v", err)
		}
		if file != nil {
			file.Close()
		}
	})

	t.Run("strips serving path prefix", func(t *testing.T) {
		m := assets.NewManager(fstest.MapFS{
			"main.js":      {Data: []byte("console.log('hello');")},
			"sub/other.js": {Data: []byte("console.log('other');")},
		}, "/assets")

		// Should be able to open with serving path prefix
		file, err := m.Open("/assets/main.js")
		if err != nil {
			t.Errorf("Expected no error when opening with serving path prefix, got %v", err)
		}
		if file != nil {
			file.Close()
		}

		file, err = m.Open("/assets/sub/other.js")
		if err != nil {
			t.Errorf("Expected no error when opening subdirectory file with prefix, got %v", err)
		}
		if file != nil {
			file.Close()
		}
	})

	t.Run("file not found", func(t *testing.T) {
		m := assets.NewManager(fstest.MapFS{
			"main.js": {Data: []byte("console.log('hello');")},
		}, "/assets")

		_, err := m.Open("nonexistent.js")
		if err == nil {
			t.Error("Expected error when opening nonexistent file")
		}
	})
}
