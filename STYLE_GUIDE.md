# LeapKit Core - Code Style Guide

This document outlines the coding standards and conventions for the LeapKit Core project to ensure consistency across all packages.

## Package Organization

### Package Documentation
Every package MUST have comprehensive package-level documentation:

```go
// Package server provides HTTP server functionality with routing capabilities,
// middleware support, and session management built on top of standard Go net/http.
// It offers a simple API for building web applications with clean routing,
// middleware chains, and error handling.
package server
```

### File Organization
- Group related functionality in the same file
- Keep files focused on a single responsibility
- Use consistent file naming patterns:
  - `manager.go` for main management logic
  - `options.go` for configuration options
  - `middleware.go` for middleware implementations
  - `*_test.go` for tests

## Naming Conventions

### Variables
- Use descriptive names for package-level variables
- Use short names for loop variables and short-lived locals
- Prefer `ctx` for context.Context
- Use `w` for http.ResponseWriter and `r` for *http.Request

### Functions and Methods
- Use verb-noun patterns: `CreateDatabase`, `HandleRequest`
- Avoid stuttering: prefer `db.Create()` over `db.CreateDatabase()`
- Use consistent naming across packages for similar operations

### Types
- Use descriptive names that indicate purpose
- Avoid generic names like `Manager` without context
- Prefer composition over inheritance

## Error Handling

### Error Messages
- Start with lowercase (Go convention)
- Use present tense: "parsing" not "failed to parse"
- Provide context without redundant "error" prefix
- Use consistent formatting:

```go
return fmt.Errorf("parsing database URL: %w", err)
return fmt.Errorf("creating database file %s: %w", u.Path, err)
```

### Error Wrapping
- Always wrap errors with context using `fmt.Errorf`
- Use `%w` verb for error wrapping
- Preserve original error information

## Documentation

### Function Comments
- Document all exported functions and methods
- Use complete sentences with proper punctuation
- Follow the format: "FunctionName does X and returns Y"
- Include parameter and return value descriptions for complex functions

```go
// Create creates a new database based on the provided URL.
// It supports both SQLite and PostgreSQL databases and returns
// an error if the database creation fails.
func Create(url string) error {
```

### Type Comments
- Document all exported types
- Explain the purpose and usage
- Include example usage when helpful

## Testing

### Test Organization
- Use `package_test` for external tests
- Use same package for internal tests when accessing unexported members
- Group related tests in the same file
- Use descriptive test names: `TestCreateDatabase_WithInvalidURL`

### Test Structure
- Follow Arrange-Act-Assert pattern
- Use table-driven tests for multiple scenarios
- Clean up resources in test cleanup functions

```go
func TestCreate(t *testing.T) {
    t.Run("valid SQLite URL", func(t *testing.T) {
        // Arrange
        tempDir := t.TempDir()
        
        // Act
        err := Create("sqlite://" + tempDir + "/test.db")
        
        // Assert
        if err != nil {
            t.Fatalf("expected no error, got %v", err)
        }
    })
}
```

## Code Structure

### Import Organization
1. Standard library imports
2. Third-party imports  
3. Internal project imports

Separate groups with blank lines:

```go
import (
    "fmt"
    "net/http"

    "github.com/gorilla/sessions"

    "go.leapkit.dev/core/server"
)
```

### Function Order
1. Package-level variables and constants
2. Types and interfaces
3. Constructor functions
4. Public methods
5. Private methods

## Interface Design

### Interface Definitions
- Keep interfaces small and focused
- Use composition over large interfaces
- Place interfaces close to their usage
- Use descriptive interface names ending in -er when appropriate

```go
type Router interface {
    Handle(pattern string, handler http.Handler)
    HandleFunc(pattern string, handler http.HandlerFunc)
    Use(middleware ...Middleware)
}
```

## Options Pattern

### Configuration
- Use functional options for configuration
- Provide sensible defaults
- Make options composable

```go
func New(options ...Option) *Server {
    s := &Server{
        host: "0.0.0.0",
        port: "3000", // sensible defaults
    }
    
    for _, option := range options {
        option(s)
    }
    
    return s
}
```

## Constants and Variables

### Package-Level Variables
- Use constants when values don't change
- Group related constants together
- Use descriptive names with proper scope

```go
const (
    DefaultHost = "0.0.0.0"
    DefaultPort = "3000"
)

var (
    // ErrInvalidURL indicates that the provided URL is malformed
    ErrInvalidURL = errors.New("invalid URL format")
)
```

## Context Usage

### Context Handling
- Always pass context as the first parameter
- Use context for cancellation and timeouts
- Store request-scoped values in context appropriately
- Don't store context in structs

## Middleware Pattern

### Middleware Implementation
- Use standard http.Handler pattern
- Chain middleware in logical order
- Provide clear middleware documentation
- Make middleware composable

```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // middleware logic
        next.ServeHTTP(w, r)
    })
}
```

## Performance Considerations

### Resource Management
- Use sync.Pool for frequently allocated objects
- Implement proper cleanup in defer statements
- Avoid unnecessary allocations in hot paths
- Use buffered channels appropriately

## Security

### Input Validation
- Validate all external inputs
- Use proper escaping for template output
- Implement rate limiting where appropriate
- Follow principle of least privilege

---

This style guide should be followed for all new code and existing code should be gradually updated to match these conventions during maintenance.