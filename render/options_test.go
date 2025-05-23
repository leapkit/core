package render

import (
	"testing"
)

func TestWithDefaultLayout(t *testing.T) {
	tests := []struct {
		name   string
		layout string
	}{
		{"standard layout", "app/layouts/main.html"},
		{"custom layout", "custom/special.html"},
		{"empty layout", ""},
		{"nested layout", "deep/nested/path/layout.html"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine(testTemplates, WithDefaultLayout(tt.layout))

			if engine.defaultLayout != tt.layout {
				t.Errorf("Expected defaultLayout %s, got %s", tt.layout, engine.defaultLayout)
			}
		})
	}
}

func TestWithHelpers(t *testing.T) {
	helpers := map[string]any{
		"uppercase":  func(s string) string { return s },
		"add":        func(a, b int) int { return a + b },
		"greet":      func(name string) string { return "Hello, " + name },
		"constant":   "CONSTANT_VALUE",
		"number":     42,
		"bool_value": true,
	}

	engine := NewEngine(testTemplates, WithHelpers(helpers))

	for key := range helpers {
		if engine.helpers[key] == nil {
			t.Errorf("Expected helper %s to be set", key)
		}
	}

	// Test that helpers don't overwrite each other
	if len(engine.helpers) != len(helpers) {
		t.Errorf("Expected %d helpers, got %d", len(helpers), len(engine.helpers))
	}
}

func TestWithHelpersEmpty(t *testing.T) {
	emptyHelpers := map[string]any{}
	engine := NewEngine(testTemplates, WithHelpers(emptyHelpers))

	if len(engine.helpers) != 0 {
		t.Errorf("Expected 0 helpers for empty map, got %d", len(engine.helpers))
	}
}

func TestWithHelpersNil(t *testing.T) {
	var nilHelpers map[string]any
	engine := NewEngine(testTemplates, WithHelpers(nilHelpers))

	if engine.helpers == nil {
		t.Error("Expected helpers map to be initialized even with nil input")
	}
}

func TestMultipleOptions(t *testing.T) {
	helpers1 := map[string]any{
		"helper1": func() string { return "result1" },
		"helper2": func() string { return "result2" },
	}

	helpers2 := map[string]any{
		"helper3": func() string { return "result3" },
		"helper4": func() string { return "result4" },
	}

	engine := NewEngine(testTemplates,
		WithDefaultLayout("custom/layout.html"),
		WithHelpers(helpers1),
		WithHelpers(helpers2),
	)

	// Check layout was set
	if engine.defaultLayout != "custom/layout.html" {
		t.Errorf("Expected custom layout, got %s", engine.defaultLayout)
	}

	// Check all helpers were set
	expectedHelpers := []string{"helper1", "helper2", "helper3", "helper4"}
	for _, helper := range expectedHelpers {
		if engine.helpers[helper] == nil {
			t.Errorf("Expected helper %s to be set", helper)
		}
	}
}

func TestWithHelpersOverride(t *testing.T) {
	helpers1 := map[string]any{
		"shared_helper": "first_value",
		"unique1":       "value1",
	}

	helpers2 := map[string]any{
		"shared_helper": "second_value",
		"unique2":       "value2",
	}

	engine := NewEngine(testTemplates,
		WithHelpers(helpers1),
		WithHelpers(helpers2),
	)

	// The second helper should override the first
	if engine.helpers["unique1"] == nil {
		t.Error("Expected unique1 helper to be preserved")
	}

	if engine.helpers["unique2"] == nil {
		t.Error("Expected unique2 helper to be set")
	}

	if engine.helpers["shared_helper"] != "second_value" {
		t.Errorf("Expected shared_helper to be overridden to 'second_value', got %v", engine.helpers["shared_helper"])
	}
}

func TestOptionOrder(t *testing.T) {
	// Test that options are applied in order
	engine := NewEngine(testTemplates,
		WithDefaultLayout("first_layout.html"),
		WithDefaultLayout("second_layout.html"),
	)

	if engine.defaultLayout != "second_layout.html" {
		t.Errorf("Expected second layout to override first, got %s", engine.defaultLayout)
	}
}

