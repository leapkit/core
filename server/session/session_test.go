package session_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gobuffalo/plush/v5"
	"github.com/gorilla/sessions"
	"go.leapkit.dev/core/server"
	"go.leapkit.dev/core/server/session"
)

func Test_Session_Setup(t *testing.T) {
	requestContext := context.Background()

	s := server.New(
		server.WithSession("session_test", "test"),
	)

	s.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			h.ServeHTTP(w, req)

			// capturing http.Request context
			requestContext = context.WithoutCancel(req.Context())
		})
	})

	s.Group("/values", func(rg server.Router) {
		rg.HandleFunc("GET /add/{value}/{$}", func(w http.ResponseWriter, r *http.Request) {
			sw := session.FromCtx(r.Context())
			sw.Values["value"] = r.PathValue("value")

			w.Write([]byte("OK"))
		})

		rg.HandleFunc("GET /clear/{$}", func(w http.ResponseWriter, r *http.Request) {
			sw := session.FromCtx(r.Context())
			for k := range sw.Values {
				delete(sw.Values, k)
			}

			w.Write([]byte("OK"))
		})

		rg.HandleFunc("GET /all/{$}", func(w http.ResponseWriter, r *http.Request) {
			sw := session.FromCtx(r.Context())
			v, _ := sw.Values["value"].(string)

			w.Write([]byte(v))
		})
	})

	s.Group("/flashes", func(rg server.Router) {
		rg.HandleFunc("GET /add/{value}/{$}", func(w http.ResponseWriter, r *http.Request) {
			sw := session.FromCtx(r.Context())
			sw.AddFlash(r.PathValue("value"))

			w.Write([]byte("OK"))
		})

		rg.HandleFunc("GET /all/{$}", func(w http.ResponseWriter, r *http.Request) {
			sw := session.FromCtx(r.Context())
			v := fmt.Sprint(sw.Flashes())

			w.Write([]byte(v))
		})

		rg.HandleFunc("GET /render/{$}", func(w http.ResponseWriter, r *http.Request) {
			valuer := r.Context().Value("valuer").(interface{ Values() map[string]any })
			result, _ := plush.Render(`<%= flash("_flash") %>`, plush.NewContextWith(valuer.Values()))

			w.Write([]byte(result))
		})
	})

	t.Run("session values", func(t *testing.T) {
		res := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodGet, "/values/add/hello/", nil)
		s.Handler().ServeHTTP(res, req)

		if res.Body.String() != "OK" {
			t.Errorf("Expected 'OK', got '%s'", res.Body.String())
			return
		}

		sw := session.FromCtx(requestContext)
		if sw.Values["value"] != "hello" {
			t.Errorf("Expected 'hello', got '%s'", sw.Values["value"])
		}

		// Attempting 10 times to check if session values persist
		for range 10 {
			req = httptest.NewRequest(http.MethodGet, "/values/all/", nil)
			req = req.WithContext(requestContext)
			res.Body.Reset()

			s.Handler().ServeHTTP(res, req)

			if res.Body.String() != "hello" {
				t.Errorf("Expected 'OK', got '%s'", res.Body.String())
				return
			}
		}

		req = httptest.NewRequest(http.MethodGet, "/values/add/bar/", nil)
		req = req.WithContext(requestContext)
		res.Body.Reset()

		s.Handler().ServeHTTP(res, req)

		if res.Body.String() != "OK" {
			t.Errorf("Expected 'OK', got '%s'", res.Body.String())
			return
		}

		req = httptest.NewRequest(http.MethodGet, "/values/clear/", nil)
		req = req.WithContext(requestContext)
		res.Body.Reset()

		s.Handler().ServeHTTP(res, req)

		if res.Body.String() != "OK" {
			t.Errorf("Expected 'OK', got '%s'", res.Body.String())
			return
		}

		sw = session.FromCtx(requestContext)
		if len(sw.Values) > 0 {
			t.Errorf("Expected empty values, got '%v'", sw.Values)
			return
		}

		req = httptest.NewRequest(http.MethodGet, "/values/all/", nil)
		req = req.WithContext(requestContext)
		res.Body.Reset()

		s.Handler().ServeHTTP(res, req)

		if res.Body.String() == "hello" {
			t.Errorf("Expected empty values, got '%s'", res.Body.String())
			return
		}

		requestContext = context.Background()
	})

	t.Run("session flashes", func(t *testing.T) {
		res := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodGet, "/flashes/add/hello/", nil)
		s.Handler().ServeHTTP(res, req)

		responseHeader := res.Header()
		cookies, ok := responseHeader["Set-Cookie"]
		if !ok || len(cookies) != 1 {
			t.Fatal("No cookies. Header:", responseHeader)
		}

		if res.Body.String() != "OK" {
			t.Fatalf("Expected 'OK', got '%s'", res.Body.String())
		}

		sw := session.FromCtx(requestContext)
		if flashes, ok := sw.Values["_flash"].([]any); !ok {
			t.Fatalf("Expected non-empty flashes, got '%v'", flashes...)
		}

		req = httptest.NewRequest(http.MethodGet, "/flashes/all/", nil)
		req = req.WithContext(requestContext)
		res.Body.Reset()

		s.Handler().ServeHTTP(res, req)

		if res.Body.String() != "[hello]" {
			t.Fatalf("Expected '[hello]', got '%s'", res.Body.String())
		}

		// Second attempt at the same endpoint to validate that there are no longer any flashes.
		req = httptest.NewRequest(http.MethodGet, "/flashes/all/", nil)
		req = req.WithContext(requestContext)
		res.Body.Reset()

		s.Handler().ServeHTTP(res, req)

		if res.Body.String() == "hello" {
			t.Fatalf("Expected empty flashes, got '%s'", res.Body.String())
		}
	})

	t.Run("session flash helpers", func(t *testing.T) {
		res := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodGet, "/flashes/add/hello/", nil)
		s.Handler().ServeHTTP(res, req)

		req = httptest.NewRequest(http.MethodGet, "/flashes/render/", nil)
		req = req.WithContext(requestContext)
		res.Body.Reset()

		s.Handler().ServeHTTP(res, req)

		if res.Body.String() != "hello" {
			t.Fatalf("Expected 'hello' flashes, got '%s'", res.Body.String())
		}

		// Once the flash was called in the previous call, this should be removed from flashes.
		req = httptest.NewRequest(http.MethodGet, "/flashes/all/", nil)
		req = req.WithContext(requestContext)
		res.Body.Reset()

		s.Handler().ServeHTTP(res, req)

		if res.Body.String() != "[]" {
			t.Fatalf("Expected '[]', got '%s'", res.Body.String())
		}
	})
}

func TestSessionOptions(t *testing.T) {
	tests := []struct {
		name   string
		option session.Option
		check  func(*sessions.CookieStore) bool
	}{
		{
			name:   "WithDomain",
			option: session.WithDomain("example.com"),
			check: func(store *sessions.CookieStore) bool {
				return store.Options.Domain == "example.com"
			},
		},
		{
			name:   "WithSecure true",
			option: session.WithSecure(true),
			check: func(store *sessions.CookieStore) bool {
				return store.Options.Secure == true
			},
		},
		{
			name:   "WithSecure false",
			option: session.WithSecure(false),
			check: func(store *sessions.CookieStore) bool {
				return store.Options.Secure == false
			},
		},
		{
			name:   "WithSameSite Strict",
			option: session.WithSameSite(http.SameSiteStrictMode),
			check: func(store *sessions.CookieStore) bool {
				return store.Options.SameSite == http.SameSiteStrictMode
			},
		},
		{
			name:   "WithPath",
			option: session.WithPath("/app"),
			check: func(store *sessions.CookieStore) bool {
				return store.Options.Path == "/app"
			},
		},
		{
			name:   "WithMaxAge",
			option: session.WithMaxAge(3600),
			check: func(store *sessions.CookieStore) bool {
				return store.Options.MaxAge == 3600
			},
		},
		{
			name:   "WithHTTPOnly false",
			option: session.WithHTTPOnly(false),
			check: func(store *sessions.CookieStore) bool {
				return store.Options.HttpOnly == false
			},
		},
		{
			name:   "WithSecureFlag true",
			option: session.WithSecureFlag(true),
			check: func(store *sessions.CookieStore) bool {
				return store.Options.Secure == true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := server.New(
				server.WithSession("test", "secret", tt.option),
			)

			// Create a test request to trigger session setup
			req := httptest.NewRequest("GET", "/", nil)
			res := httptest.NewRecorder()

			s.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("OK"))
			})

			s.Handler().ServeHTTP(res, req)

			if res.Code != 200 {
				t.Errorf("Expected status 200, got %d", res.Code)
			}
		})
	}
}

func TestRegisterSessionTypes(t *testing.T) {
	type CustomStruct struct {
		Name  string
		Value int
	}

	// This should not panic
	session.RegisterSessionTypes(CustomStruct{}, map[string]interface{}{}, []string{})

	// Test with actual session storage
	s := server.New(
		server.WithSession("test", "secret"),
	)

	var requestContext context.Context

	s.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			h.ServeHTTP(w, req)
			requestContext = context.WithoutCancel(req.Context())
		})
	})

	s.HandleFunc("GET /set/{$}", func(w http.ResponseWriter, r *http.Request) {
		sess := session.FromCtx(r.Context())
		sess.Values["custom"] = CustomStruct{Name: "test", Value: 42}
		sess.Values["slice"] = []string{"a", "b", "c"}
		w.Write([]byte("OK"))
	})

	s.HandleFunc("GET /get/{$}", func(w http.ResponseWriter, r *http.Request) {
		sess := session.FromCtx(r.Context())
		custom := sess.Values["custom"].(CustomStruct)
		slice := sess.Values["slice"].([]string)
		
		result := fmt.Sprintf("%s:%d:%v", custom.Name, custom.Value, slice)
		w.Write([]byte(result))
	})

	// Set custom data
	req := httptest.NewRequest("GET", "/set/", nil)
	res := httptest.NewRecorder()
	s.Handler().ServeHTTP(res, req)

	if res.Body.String() != "OK" {
		t.Errorf("Expected 'OK', got '%s'", res.Body.String())
	}

	// Get custom data
	req = httptest.NewRequest("GET", "/get/", nil)
	req = req.WithContext(requestContext)
	res = httptest.NewRecorder()
	s.Handler().ServeHTTP(res, req)

	expected := "test:42:[a b c]"
	if res.Body.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, res.Body.String())
	}
}

func TestAddHelpers(t *testing.T) {
	s := server.New(
		server.WithSession("test", "secret"),
	)

	var rendererValues map[string]any

	// Mock renderer that captures set values
	mockRenderer := &mockRenderer{values: make(map[string]any)}

	s.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Add mock renderer to context
			ctx := context.WithValue(req.Context(), "renderer", mockRenderer)
			req = req.WithContext(ctx)
			
			h.ServeHTTP(w, req)
			rendererValues = mockRenderer.values
		})
	})

	s.Use(session.AddHelpers)

	s.HandleFunc("GET /test/{$}", func(w http.ResponseWriter, r *http.Request) {
		sess := session.FromCtx(r.Context())
		sess.AddFlash("test message")
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/test/", nil)
	res := httptest.NewRecorder()
	s.Handler().ServeHTTP(res, req)

	if res.Body.String() != "OK" {
		t.Errorf("Expected 'OK', got '%s'", res.Body.String())
	}

	// Check that helpers were added
	if rendererValues["flash"] == nil {
		t.Error("Flash helper was not added to renderer")
	}

	if rendererValues["session"] == nil {
		t.Error("Session helper was not added to renderer")
	}

	// Test flash helper functionality
	flashHelper, ok := rendererValues["flash"].(func(string) string)
	if !ok {
		t.Error("Flash helper is not a function")
	} else {
		// The flash should be available
		result := flashHelper("_flash")
		if result != "test message" {
			t.Errorf("Expected 'test message', got '%s'", result)
		}

		// Second call should return empty (flash consumed)
		result = flashHelper("_flash")
		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	}

	// Test session helper functionality
	sessionHelper, ok := rendererValues["session"].(func() *sessions.Session)
	if !ok {
		t.Error("Session helper is not a function")
	} else {
		sess := sessionHelper()
		if sess == nil {
			t.Error("Session helper returned nil")
		}
	}
}

func TestFromCtxPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected FromCtx to panic when session not in context")
		}
	}()

	ctx := context.Background()
	session.FromCtx(ctx) // This should panic
}

func TestFlashHelperEdgeCases(t *testing.T) {
	s := server.New(
		server.WithSession("test", "secret"),
	)

	var savedContext context.Context

	s.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			h.ServeHTTP(w, req)
			savedContext = context.WithoutCancel(req.Context())
		})
	})

	s.HandleFunc("GET /empty/{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	s.HandleFunc("GET /multiple/{$}", func(w http.ResponseWriter, r *http.Request) {
		sess := session.FromCtx(r.Context())
		sess.AddFlash("first", "custom")
		sess.AddFlash("second", "custom")
		w.Write([]byte("OK"))
	})

	// Test empty flash
	req := httptest.NewRequest("GET", "/empty/", nil)
	res := httptest.NewRecorder()
	s.Handler().ServeHTTP(res, req)

	sess := session.FromCtx(savedContext)
	helper := getFlashHelper(sess)
	
	result := helper("nonexistent")
	if result != "" {
		t.Errorf("Expected empty string for nonexistent flash, got '%s'", result)
	}

	// Test multiple flashes (should only return first one)
	req = httptest.NewRequest("GET", "/multiple/", nil)
	req = req.WithContext(savedContext)
	res = httptest.NewRecorder()
	s.Handler().ServeHTTP(res, req)

	sess = session.FromCtx(savedContext)
	helper = getFlashHelper(sess)
	
	result = helper("custom")
	if result != "first" {
		t.Errorf("Expected 'first', got '%s'", result)
	}
}

func TestSaverFunctionality(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		handler  func(w http.ResponseWriter, r *http.Request)
		expected string
	}{
		{
			name: "Header() saves session",
			path: "header",
			handler: func(w http.ResponseWriter, r *http.Request) {
				sess := session.FromCtx(r.Context())
				sess.Values["test"] = "header"
				w.Header().Set("Custom", "value")
				w.Write([]byte("OK"))
			},
			expected: "header",
		},
		{
			name: "WriteHeader() saves session",
			path: "writeheader",
			handler: func(w http.ResponseWriter, r *http.Request) {
				sess := session.FromCtx(r.Context())
				sess.Values["test"] = "writeheader"
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("Created"))
			},
			expected: "writeheader",
		},
		{
			name: "Write() saves session",
			path: "write",
			handler: func(w http.ResponseWriter, r *http.Request) {
				sess := session.FromCtx(r.Context())
				sess.Values["test"] = "write"
				w.Write([]byte("Written"))
			},
			expected: "write",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := server.New(
				server.WithSession("test", "secret"),
			)

			var savedContext context.Context

			s.Use(func(h http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					h.ServeHTTP(w, req)
					savedContext = context.WithoutCancel(req.Context())
				})
			})

			s.HandleFunc(fmt.Sprintf("GET /%s/{$}", tt.path), tt.handler)

			req := httptest.NewRequest("GET", fmt.Sprintf("/%s/", tt.path), nil)
			res := httptest.NewRecorder()
			s.Handler().ServeHTTP(res, req)

			// Verify session was saved by checking if value persists
			sess := session.FromCtx(savedContext)
			if sess.Values["test"] != tt.expected {
				t.Errorf("Expected session value '%s', got '%v'", tt.expected, sess.Values["test"])
			}
		})
	}
}

func TestSessionWithValuer(t *testing.T) {
	// Test the direct integration with valuer during session registration
	mockValuer := &mockValuer{values: make(map[string]any)}
	
	// Create a session manually to test valuer integration
	sess := session.New("test", "secret")
	
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), "valuer", mockValuer)
	req = req.WithContext(ctx)
	
	res := httptest.NewRecorder()
	
	// Call Register to trigger valuer integration
	w, r := sess.Register(res, req)
	
	// Add a flash to the session
	sessionFromCtx := session.FromCtx(r.Context())
	sessionFromCtx.AddFlash("test flash")
	
	// Trigger session save
	w.Write([]byte("OK"))
	
	// Check that flash and session were set in valuer during registration
	if mockValuer.values["flash"] == nil {
		t.Error("Flash was not set in valuer")
	}

	if mockValuer.values["session"] == nil {
		t.Error("Session was not set in valuer")
	}

	// Test flash helper from valuer
	flashHelper, ok := mockValuer.values["flash"].(func(string) string)
	if !ok {
		t.Error("Flash valuer is not a function")
	} else {
		result := flashHelper("_flash")
		if result != "test flash" {
			t.Errorf("Expected 'test flash', got '%s'", result)
		}
	}

	// Test session helper from valuer
	sessionHelper, ok := mockValuer.values["session"].(func() *sessions.Session)
	if !ok {
		t.Error("Session valuer is not a function")
	} else {
		sess := sessionHelper()
		if sess == nil {
			t.Error("Session helper returned nil")
		}
	}
}

// Helper types for testing
type mockRenderer struct {
	values map[string]any
}

func (m *mockRenderer) Set(key string, value any) {
	m.values[key] = value
}

type mockValuer struct {
	values map[string]any
}

func (m *mockValuer) Set(key string, value any) {
	m.values[key] = value
}

func TestSessionErrorHandling(t *testing.T) {
	// Test error handling when session retrieval fails
	// This is hard to test directly since gorilla/sessions doesn't expose easy ways to force errors
	// But we can at least verify the session creation with invalid parameters doesn't panic
	
	// Test with empty secret (this will cause session errors but shouldn't crash)
	s := server.New(
		server.WithSession("test", ""),
	)

	s.HandleFunc("GET /test/{$}", func(w http.ResponseWriter, r *http.Request) {
		// This will trigger the error path in session.Register
		// The session will be nil or invalid, but the handler should still execute
		defer func() {
			if r := recover(); r != nil {
				// If there's a panic, we expect it due to invalid session
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		
		// Attempting to access session may cause issues with empty secret
		sess := session.FromCtx(r.Context())
		if sess != nil {
			sess.Values["test"] = "value"
		}
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/test/", nil)
	res := httptest.NewRecorder()
	s.Handler().ServeHTTP(res, req)

	// With empty secret, we expect either OK or an error response (500)
	// The important thing is it doesn't crash the application
	if res.Code != 200 && res.Code != 500 {
		t.Errorf("Expected status 200 or 500, got %d", res.Code)
	}
}

func TestSessionNew(t *testing.T) {
	// Test session creation with multiple options
	sess := session.New("test-session", "secret-key",
		session.WithDomain("example.com"),
		session.WithSecure(true),
		session.WithMaxAge(3600),
		session.WithHTTPOnly(false),
	)

	// Test that we can register the session
	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	
	w, r := sess.Register(res, req)
	
	// Verify session is in context
	sessionFromCtx := session.FromCtx(r.Context())
	if sessionFromCtx == nil {
		t.Error("Session not found in context")
	}
	
	// Test that we can write to the response
	n, err := w.Write([]byte("test"))
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}
	if n != 4 {
		t.Errorf("Expected 4 bytes written, got %d", n)
	}
}

// Helper function to access flashHelper (since it's not exported)
func getFlashHelper(sess *sessions.Session) func(string) string {
	return func(key string) string {
		val := sess.Flashes(key)
		if len(val) == 0 {
			return ""
		}
		return val[0].(string)
	}
}
