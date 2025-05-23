package render

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestPage_Set(t *testing.T) {
	engine := NewEngine(testTemplates)
	var buf bytes.Buffer
	page := engine.HTML(&buf)

	tests := []struct {
		key   string
		value any
	}{
		{"string_value", "test string"},
		{"int_value", 42},
		{"bool_value", true},
		{"slice_value", []string{"a", "b", "c"}},
		{"map_value", map[string]string{"nested": "value"}},
		{"nil_value", nil},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			page.Set(tt.key, tt.value)

			retrieved := page.Value(tt.key)
			if !reflect.DeepEqual(retrieved, tt.value) {
				t.Errorf("Expected value %v for key %s, got %v", tt.value, tt.key, retrieved)
			}
		})
	}
}

func TestPage_Value(t *testing.T) {
	engine := NewEngine(testTemplates)
	engine.Set("engine_value", "from_engine")
	
	var buf bytes.Buffer
	page := engine.HTML(&buf)
	page.Set("page_value", "from_page")

	tests := []struct {
		name     string
		key      string
		expected any
	}{
		{"engine value", "engine_value", "from_engine"},
		{"page value", "page_value", "from_page"},
		{"nonexistent value", "nonexistent", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := page.Value(tt.key)
			if result != tt.expected {
				t.Errorf("Expected %v for key %s, got %v", tt.expected, tt.key, result)
			}
		})
	}
}

func TestPage_Render(t *testing.T) {
	engine := NewEngine(testTemplates, WithDefaultLayout("app/layouts/application.html"))
	
	tests := []struct {
		name     string
		template string
		data     map[string]any
		want     string
		wantErr  bool
	}{
		{
			name:     "render with default layout",
			template: "simple.html",
			data: map[string]any{
				"title": "Test Page",
				"name":  "World",
			},
			want: "<!DOCTYPE html>\n<html>\n<head>\n    <title>Test Page</title>\n</head>\n<body>\n    <div class=\"container\">\n        Hello, World!\n    </div>\n</body>\n</html>",
			wantErr: false,
		},
		{
			name:     "render nonexistent template",
			template: "nonexistent.html",
			data:     map[string]any{},
			wantErr:  true,
		},
		{
			name:     "render with missing layout",
			template: "simple.html",
			data:     map[string]any{"name": "Test", "title": "Test Title"},
			wantErr:  false, // Should use default layout
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			page := engine.HTML(&buf)
			
			// Set template data
			for k, v := range tt.data {
				page.Set(k, v)
			}

			err := page.Render(tt.template)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			result := strings.TrimSpace(buf.String())
			if tt.want != "" && result != tt.want {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.want, result)
			}
		})
	}
}

func TestPage_RenderWithLayout(t *testing.T) {
	engine := NewEngine(testTemplates)
	
	tests := []struct {
		name     string
		template string
		layout   string
		data     map[string]any
		want     string
		wantErr  bool
	}{
		{
			name:     "render with custom layout",
			template: "simple.html",
			layout:   "custom_layout.html",
			data: map[string]any{
				"title": "Custom Test",
				"name":  "World",
			},
			want: "<!DOCTYPE html>\n<html>\n<head>\n    <title>Custom Layout - Custom Test</title>\n    <meta name=\"description\" content=\"Custom layout for testing\">\n</head>\n<body>\n    <header>\n        <h1>Custom Header</h1>\n    </header>\n    <main>\n        Hello, World!\n    </main>\n    <footer>\n        <p>Custom Footer</p>\n    </footer>\n</body>\n</html>",
			wantErr: false,
		},
		{
			name:     "render with nonexistent layout",
			template: "simple.html",
			layout:   "nonexistent_layout.html",
			data:     map[string]any{"name": "Test"},
			wantErr:  true,
		},
		{
			name:     "render nonexistent template with valid layout",
			template: "nonexistent.html",
			layout:   "custom_layout.html",
			data:     map[string]any{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			page := engine.HTML(&buf)
			
			// Set template data
			for k, v := range tt.data {
				page.Set(k, v)
			}

			err := page.RenderWithLayout(tt.template, tt.layout)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			result := strings.TrimSpace(buf.String())
			if tt.want != "" && result != tt.want {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.want, result)
			}
		})
	}
}

func TestPage_RenderClean(t *testing.T) {
	engine := NewEngine(testTemplates)
	
	tests := []struct {
		name     string
		template string
		data     map[string]any
		want     string
		wantErr  bool
	}{
		{
			name:     "render without layout",
			template: "simple.html",
			data:     map[string]any{"name": "Clean World"},
			want:     "Hello, Clean World!",
			wantErr:  false,
		},
		{
			name:     "render with helper without layout",
			template: "with_helper.html",
			data:     map[string]any{"text": "clean test"},
			want:     "CLEAN TEST",
			wantErr:  false,
		},
		{
			name:     "render nonexistent template",
			template: "nonexistent.html",
			data:     map[string]any{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			page := engine.HTML(&buf)
			page.Set("uppercase", strings.ToUpper)
			
			// Set template data
			for k, v := range tt.data {
				page.Set(k, v)
			}

			err := page.RenderClean(tt.template)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			result := strings.TrimSpace(buf.String())
			if tt.want != "" && result != tt.want {
				t.Errorf("Expected %q, got %q", tt.want, result)
			}
		})
	}
}

func TestPage_ValueOverride(t *testing.T) {
	engine := NewEngine(testTemplates)
	engine.Set("shared_key", "engine_value")
	
	var buf bytes.Buffer
	page := engine.HTML(&buf)
	
	// Page value should override engine value
	page.Set("shared_key", "page_value")
	
	result := page.Value("shared_key")
	if result != "page_value" {
		t.Errorf("Expected page value to override engine value, got %v", result)
	}
}

func TestPage_TemplateErrorHandling(t *testing.T) {
	engine := NewEngine(testTemplates)
	
	tests := []struct {
		name     string
		template string
		layout   string
		method   string
	}{
		{"render with bad template", "nonexistent.html", "", "Render"},
		{"render with bad layout", "simple.html", "bad_layout.html", "RenderWithLayout"},
		{"render clean with bad template", "nonexistent.html", "", "RenderClean"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			page := engine.HTML(&buf)
			page.Set("name", "Test")

			var err error
			switch tt.method {
			case "Render":
				err = page.Render(tt.template)
			case "RenderWithLayout":
				err = page.RenderWithLayout(tt.template, tt.layout)
			case "RenderClean":
				err = page.RenderClean(tt.template)
			}

			if err == nil {
				t.Error("Expected error for bad template/layout")
			}

			if !strings.Contains(err.Error(), "could not read file") {
				t.Errorf("Expected file read error, got: %v", err)
			}
		})
	}
}