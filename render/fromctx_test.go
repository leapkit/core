package render

import (
	"bytes"
	"context"
	"testing"
)

type testValuer struct {
	values map[string]any
}

func (tv *testValuer) Values() map[string]any {
	return tv.values
}

type invalidValuer struct{}

type complexValuer struct {
	data map[string]any
}

func (cv *complexValuer) Values() map[string]any {
	return cv.data
}

type testValuerOverride struct{}

func (tv *testValuerOverride) Values() map[string]any {
	return map[string]any{
		"shared_key": "valuer_value",
		"new_key":    "new_value",
	}
}

func TestFromCtx(t *testing.T) {
	engine := NewEngine(testTemplates)
	var buf bytes.Buffer
	page := engine.HTML(&buf)

	tests := []struct {
		name string
		ctx  context.Context
		want *Page
	}{
		{
			name: "valid context with renderer",
			ctx:  context.WithValue(context.Background(), "renderer", page),
			want: page,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromCtx(tt.ctx)
			if result != tt.want {
				t.Errorf("Expected page %v, got %v", tt.want, result)
			}
		})
	}
}

func TestFromCtxWithValuer(t *testing.T) {
	engine := NewEngine(testTemplates)
	var buf bytes.Buffer
	page := engine.HTML(&buf)

	valuer := &testValuer{
		values: map[string]any{
			"user_id":   123,
			"user_name": "testuser",
			"is_admin":  true,
		},
	}

	ctx := context.WithValue(context.Background(), "renderer", page)
	ctx = context.WithValue(ctx, "valuer", valuer)

	result := FromCtx(ctx)

	// Check that valuer values were added to the page
	if result.Value("user_id") != 123 {
		t.Errorf("Expected user_id to be 123, got %v", result.Value("user_id"))
	}

	if result.Value("user_name") != "testuser" {
		t.Errorf("Expected user_name to be 'testuser', got %v", result.Value("user_name"))
	}

	if result.Value("is_admin") != true {
		t.Errorf("Expected is_admin to be true, got %v", result.Value("is_admin"))
	}
}

func TestFromCtxWithoutValuer(t *testing.T) {
	engine := NewEngine(testTemplates)
	var buf bytes.Buffer
	page := engine.HTML(&buf)
	page.Set("existing_key", "existing_value")

	ctx := context.WithValue(context.Background(), "renderer", page)

	result := FromCtx(ctx)

	// Should return the page without modification
	if result != page {
		t.Error("Expected same page instance when no valuer present")
	}

	// Existing values should still be there
	if result.Value("existing_key") != "existing_value" {
		t.Error("Expected existing values to be preserved")
	}
}

func TestFromCtxWithInvalidValuer(t *testing.T) {
	engine := NewEngine(testTemplates)
	var buf bytes.Buffer
	page := engine.HTML(&buf)

	ctx := context.WithValue(context.Background(), "renderer", page)
	ctx = context.WithValue(ctx, "valuer", &invalidValuer{})

	result := FromCtx(ctx)

	// Should return the page without modification since valuer is invalid
	if result != page {
		t.Error("Expected same page instance when valuer is invalid")
	}
}

func TestFromCtxWithNilValuer(t *testing.T) {
	engine := NewEngine(testTemplates)
	var buf bytes.Buffer
	page := engine.HTML(&buf)

	ctx := context.WithValue(context.Background(), "renderer", page)
	ctx = context.WithValue(ctx, "valuer", nil)

	result := FromCtx(ctx)

	// Should return the page without modification
	if result != page {
		t.Error("Expected same page instance when valuer is nil")
	}
}

func TestEngineFromCtx(t *testing.T) {
	engine := NewEngine(testTemplates)

	tests := []struct {
		name string
		ctx  context.Context
		want *Engine
	}{
		{
			name: "valid context with renderEngine",
			ctx:  context.WithValue(context.Background(), "renderEngine", engine),
			want: engine,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EngineFromCtx(tt.ctx)
			if result != tt.want {
				t.Errorf("Expected engine %v, got %v", tt.want, result)
			}
		})
	}
}

func TestFromCtxPanic(t *testing.T) {
	// Test that FromCtx panics when renderer is not in context
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected FromCtx to panic when renderer not in context")
		}
	}()

	ctx := context.Background()
	FromCtx(ctx)
}

func TestEngineFromCtxPanic(t *testing.T) {
	// Test that EngineFromCtx panics when renderEngine is not in context
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected EngineFromCtx to panic when renderEngine not in context")
		}
	}()

	ctx := context.Background()
	EngineFromCtx(ctx)
}

func TestFromCtxTypePanic(t *testing.T) {
	// Test that FromCtx panics when renderer is wrong type
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected FromCtx to panic when renderer is wrong type")
		}
	}()

	ctx := context.WithValue(context.Background(), "renderer", "not_a_page")
	FromCtx(ctx)
}

func TestEngineFromCtxTypePanic(t *testing.T) {
	// Test that EngineFromCtx panics when renderEngine is wrong type
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected EngineFromCtx to panic when renderEngine is wrong type")
		}
	}()

	ctx := context.WithValue(context.Background(), "renderEngine", "not_an_engine")
	EngineFromCtx(ctx)
}

func TestFromCtxWithComplexValuer(t *testing.T) {
	engine := NewEngine(testTemplates)
	var buf bytes.Buffer
	page := engine.HTML(&buf)

	valuer := &complexValuer{
		data: map[string]any{
			"nested_map": map[string]string{"key": "value"},
			"slice":      []int{1, 2, 3},
			"function":   func() string { return "test" },
			"nil_value":  nil,
		},
	}

	ctx := context.WithValue(context.Background(), "renderer", page)
	ctx = context.WithValue(ctx, "valuer", valuer)

	result := FromCtx(ctx)

	// Check that complex values were added
	nestedMap := result.Value("nested_map")
	if nestedMap == nil {
		t.Error("Expected nested_map to be set")
	}

	slice := result.Value("slice")
	if slice == nil {
		t.Error("Expected slice to be set")
	}

	function := result.Value("function")
	if function == nil {
		t.Error("Expected function to be set")
	}

	nilValue := result.Value("nil_value")
	if nilValue != nil {
		t.Error("Expected nil_value to be nil")
	}
}

func TestFromCtxValuerOverride(t *testing.T) {
	engine := NewEngine(testTemplates)
	var buf bytes.Buffer
	page := engine.HTML(&buf)
	
	// Set an initial value in the page
	page.Set("shared_key", "page_value")

	valuer := &testValuerOverride{}

	ctx := context.WithValue(context.Background(), "renderer", page)
	ctx = context.WithValue(ctx, "valuer", valuer)

	result := FromCtx(ctx)

	// Valuer should override page values
	if result.Value("shared_key") != "valuer_value" {
		t.Errorf("Expected valuer to override page value, got %v", result.Value("shared_key"))
	}

	// New keys should be added
	if result.Value("new_key") != "new_value" {
		t.Errorf("Expected new key from valuer, got %v", result.Value("new_key"))
	}
}