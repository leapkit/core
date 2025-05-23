package hctx

import (
	"reflect"
	"testing"
)

func TestMap(t *testing.T) {
	tests := []struct {
		name     string
		input    Map
		expected map[string]interface{}
	}{
		{
			name:     "empty map",
			input:    Map{},
			expected: map[string]interface{}{},
		},
		{
			name: "map with values",
			input: Map{
				"string": "test",
				"int":    42,
				"bool":   true,
			},
			expected: map[string]interface{}{
				"string": "test",
				"int":    42,
				"bool":   true,
			},
		},
		{
			name: "map with nil value",
			input: Map{
				"nil_key": nil,
				"value":   "test",
			},
			expected: map[string]interface{}{
				"nil_key": nil,
				"value":   "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(map[string]interface{}(tt.input), tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, tt.input)
			}
		})
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name     string
		maps     []Map
		expected Map
	}{
		{
			name:     "no maps",
			maps:     []Map{},
			expected: Map{},
		},
		{
			name: "single map",
			maps: []Map{
				{"key1": "value1", "key2": "value2"},
			},
			expected: Map{"key1": "value1", "key2": "value2"},
		},
		{
			name: "two maps no overlap",
			maps: []Map{
				{"key1": "value1"},
				{"key2": "value2"},
			},
			expected: Map{"key1": "value1", "key2": "value2"},
		},
		{
			name: "two maps with overlap",
			maps: []Map{
				{"key1": "value1", "shared": "first"},
				{"key2": "value2", "shared": "second"},
			},
			expected: Map{"key1": "value1", "key2": "value2", "shared": "second"},
		},
		{
			name: "multiple maps with complex overlap",
			maps: []Map{
				{"a": 1, "b": 2, "shared": "first"},
				{"c": 3, "shared": "second"},
				{"d": 4, "shared": "third", "b": "overridden"},
			},
			expected: Map{"a": 1, "b": "overridden", "c": 3, "d": 4, "shared": "third"},
		},
		{
			name: "maps with nil values",
			maps: []Map{
				{"key1": "value1", "nil_key": nil},
				{"key2": "value2", "nil_key": "not_nil"},
			},
			expected: Map{"key1": "value1", "key2": "value2", "nil_key": "not_nil"},
		},
		{
			name: "empty maps",
			maps: []Map{
				{},
				{"key": "value"},
				{},
			},
			expected: Map{"key": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Merge(tt.maps...)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMergeComplexTypes(t *testing.T) {
	slice1 := []string{"a", "b"}
	slice2 := []string{"c", "d"}
	map1 := map[string]string{"nested": "value1"}
	map2 := map[string]string{"nested": "value2"}

	maps := []Map{
		{
			"slice": slice1,
			"map":   map1,
			"func":  func() string { return "func1" },
		},
		{
			"slice": slice2,
			"map":   map2,
			"func":  func() string { return "func2" },
		},
	}

	result := Merge(maps...)

	// Check that later values override earlier ones
	if !reflect.DeepEqual(result["slice"], slice2) {
		t.Errorf("Expected slice2 to override slice1")
	}

	if !reflect.DeepEqual(result["map"], map2) {
		t.Errorf("Expected map2 to override map1")
	}

	// Functions should also be overridden
	if result["func"] == nil {
		t.Error("Expected function to be present")
	}
}

func TestMergePreservesOriginal(t *testing.T) {
	original1 := Map{"key1": "value1", "shared": "original1"}
	original2 := Map{"key2": "value2", "shared": "original2"}

	// Create copies to check they aren't modified
	copy1 := Map{"key1": "value1", "shared": "original1"}
	copy2 := Map{"key2": "value2", "shared": "original2"}

	result := Merge(original1, original2)

	// Check that originals weren't modified
	if !reflect.DeepEqual(original1, copy1) {
		t.Error("Original map 1 was modified")
	}

	if !reflect.DeepEqual(original2, copy2) {
		t.Error("Original map 2 was modified")
	}

	// Check that result is correct
	expected := Map{"key1": "value1", "key2": "value2", "shared": "original2"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestMergeLargeNumberOfMaps(t *testing.T) {
	var maps []Map
	expected := Map{}

	// Create 100 maps with unique keys
	for i := 0; i < 100; i++ {
		key := string(rune('a' + i%26)) // a-z repeating
		value := i
		maps = append(maps, Map{key: value})
		expected[key] = value // Later values will override
	}

	result := Merge(maps...)

	if len(result) != len(expected) {
		t.Errorf("Expected %d keys, got %d", len(expected), len(result))
	}

	for k, v := range expected {
		if result[k] != v {
			t.Errorf("Expected %v for key %s, got %v", v, k, result[k])
		}
	}
}

func TestMergeWithNilMap(t *testing.T) {
	map1 := Map{"key1": "value1"}
	var nilMap Map
	map2 := Map{"key2": "value2"}

	result := Merge(map1, nilMap, map2)

	expected := Map{"key1": "value1", "key2": "value2"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}