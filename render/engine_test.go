package render

import (
	"bytes"
	"reflect"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
)

var testTemplates = fstest.MapFS{
	"simple.html": &fstest.MapFile{
		Data: []byte("Hello, <%= name %>!"),
	},
	"with_globals.html": &fstest.MapFile{
		Data: []byte("<%= global_title %> - <%= subtitle %>"),
	},
	"with_helper.html": &fstest.MapFile{
		Data: []byte("<%= uppercase(text) %>"),
	},
	"app/layouts/application.html": &fstest.MapFile{
		Data: []byte(`<!DOCTYPE html>
<html>
<head>
    <title><%= title %></title>
</head>
<body>
    <div class="container">
        <%= yield %>
    </div>
</body>
</html>`),
	},
	"custom_layout.html": &fstest.MapFile{
		Data: []byte(`<!DOCTYPE html>
<html>
<head>
    <title>Custom Layout - <%= title %></title>
    <meta name="description" content="Custom layout for testing">
</head>
<body>
    <header>
        <h1>Custom Header</h1>
    </header>
    <main>
        <%= yield %>
    </main>
    <footer>
        <p>Custom Footer</p>
    </footer>
</body>
</html>`),
	},
}

func TestNewEngine(t *testing.T) {
	tests := []struct {
		name           string
		fs             fstest.MapFS
		options        []Option
		expectedLayout string
	}{
		{
			name:           "basic engine creation",
			fs:             testTemplates,
			expectedLayout: "app/layouts/application.html",
		},
		{
			name: "engine with custom layout option",
			fs:   testTemplates,
			options: []Option{
				WithDefaultLayout("custom/layout.html"),
			},
			expectedLayout: "custom/layout.html",
		},
		{
			name: "engine with helpers option",
			fs:   testTemplates,
			options: []Option{
				WithHelpers(map[string]any{
					"uppercase": strings.ToUpper,
					"greet":     func(name string) string { return "Hello, " + name },
				}),
			},
			expectedLayout: "app/layouts/application.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine(tt.fs, tt.options...)

			if engine.templates == nil {
				t.Error("Expected templates to be set")
			}

			if engine.defaultLayout != tt.expectedLayout {
				t.Errorf("Expected defaultLayout %s, got %s", tt.expectedLayout, engine.defaultLayout)
			}

			if engine.values == nil {
				t.Error("Expected values map to be initialized")
			}

			if engine.helpers == nil {
				t.Error("Expected helpers map to be initialized")
			}
		})
	}
}

func TestEngine_Set(t *testing.T) {
	engine := NewEngine(testTemplates)

	tests := []struct {
		key   string
		value any
	}{
		{"string_key", "string_value"},
		{"int_key", 42},
		{"bool_key", true},
		{"slice_key", []string{"a", "b", "c"}},
		{"map_key", map[string]string{"nested": "value"}},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			engine.Set(tt.key, tt.value)

			if !reflect.DeepEqual(engine.values[tt.key], tt.value) {
				t.Errorf("Expected value %v for key %s, got %v", tt.value, tt.key, engine.values[tt.key])
			}
		})
	}
}

func TestEngine_SetHelper(t *testing.T) {
	engine := NewEngine(testTemplates)

	upperFunc := strings.ToUpper
	greetFunc := func(name string) string { return "Hello, " + name }

	tests := []struct {
		key   string
		value any
	}{
		{"uppercase", upperFunc},
		{"greet", greetFunc},
		{"constant", "CONSTANT_VALUE"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			engine.SetHelper(tt.key, tt.value)

			if engine.helpers[tt.key] == nil {
				t.Errorf("Expected helper %s to be set", tt.key)
			}
		})
	}
}

func TestEngine_HTML(t *testing.T) {
	engine := NewEngine(testTemplates)
	engine.Set("global_value", "test")
	engine.SetHelper("test_helper", func() string { return "helper_result" })

	var buf bytes.Buffer
	page := engine.HTML(&buf)

	if page == nil {
		t.Fatal("Expected page to be created")
	}

	if page.fs == nil {
		t.Error("Expected page to have filesystem set")
	}

	if page.writer != &buf {
		t.Error("Expected page to have correct writer")
	}

	if page.defaultLayout != engine.defaultLayout {
		t.Error("Expected page to have same default layout as engine")
	}

	if page.context == nil {
		t.Error("Expected page context to be initialized")
	}

	// Test that global values are set in context
	if page.Value("global_value") != "test" {
		t.Error("Expected global values to be copied to page context")
	}
}

func TestEngine_RenderHTML(t *testing.T) {
	engine := NewEngine(testTemplates)
	engine.Set("global_title", "Global Title")
	engine.SetHelper("uppercase", strings.ToUpper)

	tests := []struct {
		name     string
		template string
		values   map[string]any
		want     string
		wantErr  bool
	}{
		{
			name:     "simple template",
			template: "simple.html",
			values:   map[string]any{"name": "World"},
			want:     "Hello, World!",
			wantErr:  false,
		},
		{
			name:     "template with global values",
			template: "with_globals.html",
			values:   map[string]any{"subtitle": "Local Subtitle"},
			want:     "Global Title - Local Subtitle",
			wantErr:  false,
		},
		{
			name:     "template with helpers",
			template: "with_helper.html",
			values:   map[string]any{"text": "hello world"},
			want:     "HELLO WORLD",
			wantErr:  false,
		},
		{
			name:     "nonexistent template",
			template: "nonexistent.html",
			values:   map[string]any{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.RenderHTML(tt.template, tt.values)

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

			if strings.TrimSpace(result) != tt.want {
				t.Errorf("Expected %q, got %q", tt.want, strings.TrimSpace(result))
			}
		})
	}
}

func TestEngine_ConcurrentAccess(t *testing.T) {
	engine := NewEngine(testTemplates)

	var wg sync.WaitGroup
	numGoroutines := 100

	// Test concurrent Set operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			engine.Set("key", index)
		}(i)
	}

	// Test concurrent SetHelper operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			engine.SetHelper("helper", func() int { return index })
		}(i)
	}

	// Test concurrent HTML page creation
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var buf bytes.Buffer
			page := engine.HTML(&buf)
			if page == nil {
				t.Error("Expected page to be created")
			}
		}()
	}

	wg.Wait()

	// Verify that the engine state is still valid
	if engine.values == nil {
		t.Error("Engine values map should not be nil after concurrent access")
	}

	if engine.helpers == nil {
		t.Error("Engine helpers map should not be nil after concurrent access")
	}
}

func TestEngine_OptionsIntegration(t *testing.T) {
	customHelpers := map[string]any{
		"multiply": func(a, b int) int { return a * b },
		"concat":   func(a, b string) string { return a + b },
	}

	engine := NewEngine(testTemplates,
		WithDefaultLayout("custom/layout.html"),
		WithHelpers(customHelpers),
	)

	if engine.defaultLayout != "custom/layout.html" {
		t.Errorf("Expected custom layout, got %s", engine.defaultLayout)
	}

	for key := range customHelpers {
		if engine.helpers[key] == nil {
			t.Errorf("Expected helper %s to be set", key)
		}
	}

	// Test that helpers work in HTML pages
	var buf bytes.Buffer
	page := engine.HTML(&buf)

	if page.Value("multiply") == nil {
		t.Error("Expected helper to be available in page context")
	}
}