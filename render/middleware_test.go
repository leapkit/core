package render

import (
	"context"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestMiddleware(t *testing.T) {
	// Create a simple test filesystem
	testFS := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte("Hello, <%= name %>!"),
		},
		"app/layouts/application.html": &fstest.MapFile{
			Data: []byte("Layout: <%= yield %>"),
		},
	}

	tests := []struct {
		name    string
		fs      fs.FS
		options []Option
	}{
		{
			name:    "basic middleware",
			fs:      testFS,
			options: nil,
		},
		{
			name: "middleware with options",
			fs:   testFS,
			options: []Option{
				WithDefaultLayout("custom/layout.html"),
				WithHelpers(map[string]any{
					"test_helper": func() string { return "helper_result" },
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := Middleware(tt.fs, tt.options...)

			// Create a test handler that checks if the render context is available
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check if renderer is in context
				renderer := r.Context().Value("renderer")
				if renderer == nil {
					t.Error("Expected renderer to be in context")
					return
				}

				page, ok := renderer.(*Page)
				if !ok {
					t.Error("Expected renderer to be a *Page")
					return
				}

				if page == nil {
					t.Error("Expected page to not be nil")
					return
				}

				// Check if renderEngine is in context
				engine := r.Context().Value("renderEngine")
				if engine == nil {
					t.Error("Expected renderEngine to be in context")
					return
				}

				renderEngine, ok := engine.(*Engine)
				if !ok {
					t.Error("Expected renderEngine to be an *Engine")
					return
				}

				if renderEngine == nil {
					t.Error("Expected renderEngine to not be nil")
					return
				}

				w.WriteHeader(http.StatusOK)
			})

			// Wrap the handler with middleware
			wrappedHandler := middleware(handler)

			// Create test request
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			// Execute the request
			wrappedHandler.ServeHTTP(w, req)

			// Check response status
			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		})
	}
}

func TestInCtx(t *testing.T) {
	testFS := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte("Test content"),
		},
	}

	// Test that InCtx is an alias for Middleware
	middleware1 := InCtx(testFS)
	middleware2 := Middleware(testFS)

	// Both should create valid middleware functions
	if middleware1 == nil {
		t.Error("Expected InCtx to return a middleware function")
	}

	if middleware2 == nil {
		t.Error("Expected Middleware to return a middleware function")
	}

	// Test that both work the same way
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renderer1 := r.Context().Value("renderer")
		engine1 := r.Context().Value("renderEngine")

		if renderer1 == nil || engine1 == nil {
			t.Error("Expected context values to be set")
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()

	middleware1(handler).ServeHTTP(w1, req)
	middleware2(handler).ServeHTTP(w2, req.Clone(context.Background()))

	if w1.Code != w2.Code {
		t.Error("Expected InCtx and Middleware to behave identically")
	}
}

func TestMiddlewareWithComplexOptions(t *testing.T) {
	testFS := fstest.MapFS{
		"page.html": &fstest.MapFile{
			Data: []byte("Hello, <%= greet(name) %>!"),
		},
		"custom/layout.html": &fstest.MapFile{
			Data: []byte("Custom: <%= yield %>"),
		},
	}

	helpers := map[string]any{
		"greet": func(name string) string { return "Mr. " + name },
		"upper": func(s string) string { return s },
	}

	middleware := Middleware(testFS,
		WithDefaultLayout("custom/layout.html"),
		WithHelpers(helpers),
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.Context().Value("renderer").(*Page)
		engine := r.Context().Value("renderEngine").(*Engine)

		// Check that the engine has the correct default layout
		if engine.defaultLayout != "custom/layout.html" {
			t.Errorf("Expected custom layout, got %s", engine.defaultLayout)
		}

		// Check that helpers are available in the page
		if page.Value("greet") == nil {
			t.Error("Expected greet helper to be available in page")
		}

		if page.Value("upper") == nil {
			t.Error("Expected upper helper to be available in page")
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware(handler).ServeHTTP(w, req)
}

func TestMiddlewareContextIsolation(t *testing.T) {
	testFS := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte("Test"),
		},
	}

	middleware := Middleware(testFS)

	// Test that each request gets its own context values
	var page1, page2 *Page
	var engine1, engine2 *Engine

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/first" {
			page1 = r.Context().Value("renderer").(*Page)
			engine1 = r.Context().Value("renderEngine").(*Engine)
		} else {
			page2 = r.Context().Value("renderer").(*Page)
			engine2 = r.Context().Value("renderEngine").(*Engine)
		}
	})

	wrappedHandler := middleware(handler)

	// First request
	req1 := httptest.NewRequest("GET", "/first", nil)
	w1 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w1, req1)

	// Second request
	req2 := httptest.NewRequest("GET", "/second", nil)
	w2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w2, req2)

	// Pages should be different instances
	if page1 == page2 {
		t.Error("Expected different page instances for different requests")
	}

	// Engines should be the same instance (shared)
	if engine1 != engine2 {
		t.Error("Expected same engine instance for different requests")
	}

	// But pages should have different writers
	if page1.writer == page2.writer {
		t.Error("Expected different writers for different page instances")
	}
}

func TestMiddlewareChaining(t *testing.T) {
	testFS := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte("Test"),
		},
	}

	renderMiddleware := Middleware(testFS)

	// Create another middleware that sets a context value
	customMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "custom_key", "custom_value")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that both middleware values are present
		renderer := r.Context().Value("renderer")
		engine := r.Context().Value("renderEngine")
		custom := r.Context().Value("custom_key")

		if renderer == nil {
			t.Error("Expected renderer in context")
		}
		if engine == nil {
			t.Error("Expected renderEngine in context")
		}
		if custom != "custom_value" {
			t.Error("Expected custom value in context")
		}
	})

	// Chain middlewares
	finalHandler := renderMiddleware(customMiddleware(handler))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	finalHandler.ServeHTTP(w, req)
}

func TestMiddlewareWithEmptyFS(t *testing.T) {
	emptyFS := fstest.MapFS{}

	middleware := Middleware(emptyFS)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renderer := r.Context().Value("renderer")
		engine := r.Context().Value("renderEngine")

		if renderer == nil {
			t.Error("Expected renderer even with empty FS")
		}
		if engine == nil {
			t.Error("Expected engine even with empty FS")
		}

		// Should still be able to create page and engine
		page, ok := renderer.(*Page)
		if !ok || page == nil {
			t.Error("Expected valid page instance")
		}

		renderEngine, ok := engine.(*Engine)
		if !ok || renderEngine == nil {
			t.Error("Expected valid engine instance")
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware(handler).ServeHTTP(w, req)
}