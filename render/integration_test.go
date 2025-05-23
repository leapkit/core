package render

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"testing/fstest"
)

func TestIntegrationRenderWithErrors(t *testing.T) {
	// Test with template that has syntax errors
	badTemplateFS := fstest.MapFS{
		"bad_syntax.html": &fstest.MapFile{
			Data: []byte("<%= unclosed"),
		},
		"app/layouts/application.html": &fstest.MapFile{
			Data: []byte("Layout: <%= yield %>"),
		},
	}

	engine := NewEngine(badTemplateFS)
	var buf bytes.Buffer
	page := engine.HTML(&buf)

	err := page.Render("bad_syntax.html")
	if err == nil {
		t.Error("Expected error for bad template syntax")
	}
}

func TestIntegrationRenderHTMLErrors(t *testing.T) {
	// Test RenderHTML with template syntax errors
	badTemplateFS := fstest.MapFS{
		"bad_syntax.html": &fstest.MapFile{
			Data: []byte("<%= unclosed"),
		},
	}

	engine := NewEngine(badTemplateFS)
	_, err := engine.RenderHTML("bad_syntax.html", map[string]any{})
	if err == nil {
		t.Error("Expected error for bad template syntax in RenderHTML")
	}
}

func TestIntegrationLayoutWithSyntaxError(t *testing.T) {
	// Test with layout that has syntax errors
	badLayoutFS := fstest.MapFS{
		"page.html": &fstest.MapFile{
			Data: []byte("Page content"),
		},
		"app/layouts/application.html": &fstest.MapFile{
			Data: []byte("<%= unclosed layout"),
		},
	}

	engine := NewEngine(badLayoutFS)
	var buf bytes.Buffer
	page := engine.HTML(&buf)

	err := page.Render("page.html")
	if err == nil {
		t.Error("Expected error for bad layout syntax")
	}
}

func TestIntegrationWriteError(t *testing.T) {
	// Test write error by using a writer that always fails
	engine := NewEngine(testTemplates)
	failWriter := &failingWriter{}
	page := engine.HTML(failWriter)
	page.Set("name", "Test")

	err := page.RenderClean("simple.html")
	if err == nil {
		t.Error("Expected write error")
	}

	if !strings.Contains(err.Error(), "could not write to response") {
		t.Errorf("Expected write error message, got: %v", err)
	}
}

func TestIntegrationHelperInLayout(t *testing.T) {
	// Test using helpers in layout
	helperFS := fstest.MapFS{
		"page.html": &fstest.MapFile{
			Data: []byte("Content: <%= name %>"),
		},
		"app/layouts/application.html": &fstest.MapFile{
			Data: []byte("Title: <%= uppercase(title) %>\nBody: <%= yield %>"),
		},
	}

	engine := NewEngine(helperFS)
	engine.SetHelper("uppercase", strings.ToUpper)

	var buf bytes.Buffer
	page := engine.HTML(&buf)
	page.Set("name", "World")
	page.Set("title", "test page")

	err := page.Render("page.html")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	result := buf.String()
	if !strings.Contains(result, "Title: TEST PAGE") {
		t.Errorf("Expected helper to work in layout, got: %s", result)
	}
	if !strings.Contains(result, "Content: World") {
		t.Errorf("Expected page content, got: %s", result)
	}
}

func TestIntegrationComplexTemplateStructure(t *testing.T) {
	// Test complex nested template structure
	complexFS := fstest.MapFS{
		"nested/deep/page.html": &fstest.MapFile{
			Data: []byte("Deep page: <%= data %>"),
		},
		"layouts/custom.html": &fstest.MapFile{
			Data: []byte("Custom: <%= yield %>"),
		},
	}

	engine := NewEngine(complexFS, WithDefaultLayout("layouts/custom.html"))
	var buf bytes.Buffer
	page := engine.HTML(&buf)
	page.Set("data", "nested content")

	err := page.Render("nested/deep/page.html")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	result := strings.TrimSpace(buf.String())
	expected := "Custom: Deep page: nested content"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestIntegrationRenderWithLayoutErrors(t *testing.T) {
	// Test RenderWithLayout with various error conditions
	engine := NewEngine(testTemplates)

	tests := []struct {
		name     string
		template string
		layout   string
		wantErr  string
	}{
		{
			name:     "bad template in RenderWithLayout",
			template: "nonexistent.html",
			layout:   "custom_layout.html",
			wantErr:  "could not read file",
		},
		{
			name:     "bad layout in RenderWithLayout",
			template: "simple.html",
			layout:   "nonexistent_layout.html",
			wantErr:  "could not read file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			page := engine.HTML(&buf)
			page.Set("name", "Test")
			page.Set("title", "Test Title")

			err := page.RenderWithLayout(tt.template, tt.layout)
			if err == nil {
				t.Error("Expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Expected error containing %q, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestIntegrationPartialFeeder(t *testing.T) {
	// Test the partialFeeder helper function
	partialFS := fstest.MapFS{
		"partial.html": &fstest.MapFile{
			Data: []byte("Partial content: test"),
		},
		"main.html": &fstest.MapFile{
			Data: []byte("Main: <%= partialFeeder(\"partial.html\") %>"),
		},
		"app/layouts/application.html": &fstest.MapFile{
			Data: []byte("<%= yield %>"),
		},
	}

	engine := NewEngine(partialFS)
	var buf bytes.Buffer
	page := engine.HTML(&buf)

	err := page.RenderClean("main.html")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "Main: Partial content: test"
	if strings.TrimSpace(buf.String()) != expected {
		t.Errorf("Expected %q, got %q", expected, strings.TrimSpace(buf.String()))
	}
}

// failingWriter always returns an error on Write
type failingWriter struct{}

func (fw *failingWriter) Write(p []byte) (int, error) {
	return 0, io.ErrShortWrite
}