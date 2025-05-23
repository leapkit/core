package render

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestTemplateFS(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create some test files in the temp directory
	testFile := filepath.Join(tempDir, "test.html")
	err := os.WriteFile(testFile, []byte("local file content"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	embedFS := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte("embed content"),
		},
	}

	tests := []struct {
		name     string
		embed    fstest.MapFS
		dir      string
		useLocal bool
	}{
		{
			name:     "with directory",
			embed:    embedFS,
			dir:      tempDir,
			useLocal: true,
		},
		{
			name:     "empty directory defaults to pwd",
			embed:    embedFS,
			dir:      "",
			useLocal: true,
		},
		{
			name:     "production mode uses embed",
			embed:    embedFS,
			dir:      tempDir,
			useLocal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment for test
			originalEnv := os.Getenv("GO_ENV")
			defer os.Setenv("GO_ENV", originalEnv)

			if tt.useLocal {
				os.Setenv("GO_ENV", "development")
			} else {
				os.Setenv("GO_ENV", "production")
			}

			fs := TemplateFS(tt.embed, tt.dir)

			if fs.embed == nil {
				t.Error("Expected embed FS to be set")
			}

			if tt.dir != "" && fs.dir != tt.dir {
				t.Errorf("Expected dir %s, got %s", tt.dir, fs.dir)
			}

			if tt.dir == "" && fs.dir == "" {
				t.Error("Expected dir to be set to current working directory when empty")
			}

			if fs.useLocal != tt.useLocal {
				t.Errorf("Expected useLocal %v, got %v", tt.useLocal, fs.useLocal)
			}
		})
	}
}

func TestTemplateFS_Open_LocalFile(t *testing.T) {
	// Create a temporary directory with test files
	tempDir := t.TempDir()
	localFile := filepath.Join(tempDir, "local.html")
	err := os.WriteFile(localFile, []byte("local content"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create local test file: %v", err)
	}

	// Create embed FS with different content
	embedFS := fstest.MapFS{
		"local.html": &fstest.MapFile{
			Data: []byte("embed content"),
		},
	}

	// Set to development mode to prefer local files
	originalEnv := os.Getenv("GO_ENV")
	defer os.Setenv("GO_ENV", originalEnv)
	os.Setenv("GO_ENV", "development")

	fs := TemplateFS(embedFS, tempDir)

	file, err := fs.Open("local.html")
	if err != nil {
		t.Fatalf("Failed to open local file: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read local file: %v", err)
	}

	if string(content) != "local content" {
		t.Errorf("Expected local content, got %s", string(content))
	}
}

func TestTemplateFS_Open_EmbedFallback(t *testing.T) {
	// Create a temporary directory without the file
	tempDir := t.TempDir()

	// Create embed FS with the file
	embedFS := fstest.MapFS{
		"embed_only.html": &fstest.MapFile{
			Data: []byte("embed content"),
		},
	}

	// Set to development mode but file doesn't exist locally
	originalEnv := os.Getenv("GO_ENV")
	defer os.Setenv("GO_ENV", originalEnv)
	os.Setenv("GO_ENV", "development")

	fs := TemplateFS(embedFS, tempDir)

	file, err := fs.Open("embed_only.html")
	if err != nil {
		t.Fatalf("Failed to open embed file: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read embed file: %v", err)
	}

	if string(content) != "embed content" {
		t.Errorf("Expected embed content, got %s", string(content))
	}
}

func TestTemplateFS_Open_ProductionMode(t *testing.T) {
	// Create a temporary directory with test files
	tempDir := t.TempDir()
	localFile := filepath.Join(tempDir, "test.html")
	err := os.WriteFile(localFile, []byte("local content"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create local test file: %v", err)
	}

	// Create embed FS with different content
	embedFS := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte("embed content"),
		},
	}

	// Set to production mode to use embed FS
	originalEnv := os.Getenv("GO_ENV")
	defer os.Setenv("GO_ENV", originalEnv)
	os.Setenv("GO_ENV", "production")

	fs := TemplateFS(embedFS, tempDir)

	file, err := fs.Open("test.html")
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Should get embed content even though local file exists
	if string(content) != "embed content" {
		t.Errorf("Expected embed content in production, got %s", string(content))
	}
}

func TestTemplateFS_Open_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	embedFS := fstest.MapFS{}

	fs := TemplateFS(embedFS, tempDir)

	_, err := fs.Open("nonexistent.html")
	if err == nil {
		t.Error("Expected error when opening nonexistent file")
	}
}

func TestTemplateFS_ReadFile(t *testing.T) {
	// Create a temporary directory with test files
	tempDir := t.TempDir()
	localFile := filepath.Join(tempDir, "readme.txt")
	localContent := "This is local readme content"
	err := os.WriteFile(localFile, []byte(localContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create local test file: %v", err)
	}

	// Create embed FS
	embedFS := fstest.MapFS{
		"readme.txt": &fstest.MapFile{
			Data: []byte("This is embed readme content"),
		},
		"embed_only.txt": &fstest.MapFile{
			Data: []byte("Only in embed"),
		},
	}

	tests := []struct {
		name     string
		filename string
		env      string
		expected string
		wantErr  bool
	}{
		{
			name:     "local file in development",
			filename: "readme.txt",
			env:      "development",
			expected: "This is local readme content",
			wantErr:  false,
		},
		{
			name:     "embed file in production",
			filename: "readme.txt",
			env:      "production",
			expected: "This is embed readme content",
			wantErr:  false,
		},
		{
			name:     "embed only file",
			filename: "embed_only.txt",
			env:      "development",
			expected: "Only in embed",
			wantErr:  false,
		},
		{
			name:     "nonexistent file",
			filename: "nonexistent.txt",
			env:      "development",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalEnv := os.Getenv("GO_ENV")
			defer os.Setenv("GO_ENV", originalEnv)
			os.Setenv("GO_ENV", tt.env)

			fs := TemplateFS(embedFS, tempDir)

			content, err := fs.ReadFile(tt.filename)

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

			if string(content) != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, string(content))
			}
		})
	}
}

func TestTemplateFS_EnvironmentDefault(t *testing.T) {
	// Test default environment behavior
	originalEnv := os.Getenv("GO_ENV")
	defer os.Setenv("GO_ENV", originalEnv)
	os.Unsetenv("GO_ENV")

	tempDir := t.TempDir()
	embedFS := fstest.MapFS{}

	fs := TemplateFS(embedFS, tempDir)

	// Default should be development
	if !fs.useLocal {
		t.Error("Expected useLocal to be true when GO_ENV is not set (should default to development)")
	}
}

func TestTemplateFS_EmptyDir(t *testing.T) {
	embedFS := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte("test content"),
		},
	}

	fs := TemplateFS(embedFS, "")

	// Should set dir to current working directory
	if fs.dir == "" {
		t.Error("Expected dir to be set when empty string passed")
	}

	// Should still be able to open files from embed
	file, err := fs.Open("test.html")
	if err != nil {
		t.Fatalf("Failed to open file from embed: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != "test content" {
		t.Errorf("Expected test content, got %s", string(content))
	}
}

func TestTemplateFS_DirectoryStructure(t *testing.T) {
	// Create nested directory structure
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "nested", "deep")
	err := os.MkdirAll(nestedDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	nestedFile := filepath.Join(nestedDir, "deep.html")
	err = os.WriteFile(nestedFile, []byte("deep content"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create nested file: %v", err)
	}

	embedFS := fstest.MapFS{
		"nested/deep/deep.html": &fstest.MapFile{
			Data: []byte("embed deep content"),
		},
	}

	originalEnv := os.Getenv("GO_ENV")
	defer os.Setenv("GO_ENV", originalEnv)
	os.Setenv("GO_ENV", "development")

	fs := TemplateFS(embedFS, tempDir)

	file, err := fs.Open("nested/deep/deep.html")
	if err != nil {
		t.Fatalf("Failed to open nested file: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read nested file: %v", err)
	}

	// Should get local content in development mode
	if string(content) != "deep content" {
		t.Errorf("Expected deep content, got %s", string(content))
	}
}
