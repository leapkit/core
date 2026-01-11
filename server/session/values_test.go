package session_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.leapkit.dev/core/server"
	"go.leapkit.dev/core/server/session"
)

func TestGet(t *testing.T) {
	t.Run("returns value when key exists", func(t *testing.T) {
		var savedCtx context.Context

		s := server.New(
			server.WithSession("secret", "test"),
		)

		s.Use(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				h.ServeHTTP(w, r)
				savedCtx = context.WithoutCancel(r.Context())
			})
		})

		s.HandleFunc("GET /set/{$}", func(w http.ResponseWriter, r *http.Request) {
			err := session.Set(r, w, "user_id", "1234")
			if err != nil {
				t.Errorf("Set failed: %v", err)
			}
			w.Write([]byte("OK"))
		})

		s.HandleFunc("GET /get/{$}", func(w http.ResponseWriter, r *http.Request) {
			val, exists, err := session.Get(r, "user_id")
			if err != nil {
				t.Errorf("Get failed: %v", err)
			}
			if !exists {
				t.Error("Expected key to exist")
			}
			if val != "1234" {
				t.Errorf("Expected '1234', got '%v'", val)
			}
			w.Write([]byte("OK"))
		})

		// Set value
		req := httptest.NewRequest("GET", "/set/", nil)
		res := httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)

		// Get value
		req = httptest.NewRequest("GET", "/get/", nil)
		req = req.WithContext(savedCtx)
		res = httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)
	})

	t.Run("returns false when key does not exist", func(t *testing.T) {
		s := server.New(
			server.WithSession("secret", "test"),
		)

		s.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
			val, exists, err := session.Get(r, "nonexistent")
			if err != nil {
				t.Errorf("Get failed: %v", err)
			}
			if exists {
				t.Error("Expected key to not exist")
			}
			if val != nil {
				t.Errorf("Expected nil, got '%v'", val)
			}
			w.Write([]byte("OK"))
		})

		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)
	})

	t.Run("returns error when session not in context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)

		_, _, err := session.Get(req, "key")
		if !errors.Is(err, session.ErrSessionNotFound) {
			t.Errorf("Expected ErrSessionNotFound, got %v", err)
		}
	})
}

func TestSet(t *testing.T) {
	t.Run("sets value successfully", func(t *testing.T) {
		s := server.New(
			server.WithSession("secret", "test"),
		)

		s.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
			err := session.Set(r, w, "user_id", "1234")
			if err != nil {
				t.Errorf("Set failed: %v", err)
			}

			// Verify with Get
			val, exists, err := session.Get(r, "user_id")
			if err != nil {
				t.Errorf("Get failed: %v", err)
			}
			if !exists {
				t.Error("Expected key to exist after Set")
			}
			if val != "1234" {
				t.Errorf("Expected '1234', got '%v'", val)
			}
			w.Write([]byte("OK"))
		})

		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)
	})

	t.Run("sets different value types", func(t *testing.T) {
		s := server.New(
			server.WithSession("secret", "test"),
		)

		s.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
			testCases := []struct {
				key   string
				value any
			}{
				{"string", "hello"},
				{"int", 42},
				{"bool", true},
				{"slice", []string{"a", "b"}},
			}

			for _, tc := range testCases {
				err := session.Set(r, w, tc.key, tc.value)
				if err != nil {
					t.Errorf("Set failed for %s: %v", tc.key, err)
				}
			}
			w.Write([]byte("OK"))
		})

		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)
	})

	t.Run("returns error when session not in context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()

		err := session.Set(req, res, "key", "value")
		if !errors.Is(err, session.ErrSessionNotFound) {
			t.Errorf("Expected ErrSessionNotFound, got %v", err)
		}
	})
}

func TestDelete(t *testing.T) {
	t.Run("deletes existing key", func(t *testing.T) {
		s := server.New(
			server.WithSession("secret", "test"),
		)

		s.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
			// Set a value
			err := session.Set(r, w, "user_id", "1234")
			if err != nil {
				t.Errorf("Set failed: %v", err)
			}

			// Delete it
			err = session.Delete(r, w, "user_id")
			if err != nil {
				t.Errorf("Delete failed: %v", err)
			}

			// Verify it's gone
			_, exists, err := session.Get(r, "user_id")
			if err != nil {
				t.Errorf("Get failed: %v", err)
			}
			if exists {
				t.Error("Expected key to not exist after Delete")
			}
			w.Write([]byte("OK"))
		})

		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)
	})

	t.Run("does not error when key does not exist", func(t *testing.T) {
		s := server.New(
			server.WithSession("secret", "test"),
		)

		s.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
			err := session.Delete(r, w, "nonexistent")
			if err != nil {
				t.Errorf("Delete should not fail for nonexistent key: %v", err)
			}
			w.Write([]byte("OK"))
		})

		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)
	})

	t.Run("returns error when session not in context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()

		err := session.Delete(req, res, "key")
		if !errors.Is(err, session.ErrSessionNotFound) {
			t.Errorf("Expected ErrSessionNotFound, got %v", err)
		}
	})
}

func TestClear(t *testing.T) {
	t.Run("clears all values", func(t *testing.T) {
		s := server.New(
			server.WithSession("secret", "test"),
		)

		s.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
			// Set multiple values
			session.Set(r, w, "key1", "value1")
			session.Set(r, w, "key2", "value2")
			session.Set(r, w, "key3", "value3")

			// Clear all
			err := session.Clear(r, w)
			if err != nil {
				t.Errorf("Clear failed: %v", err)
			}

			// Verify all are gone
			for _, key := range []string{"key1", "key2", "key3"} {
				_, exists, _ := session.Get(r, key)
				if exists {
					t.Errorf("Expected key %s to not exist after Clear", key)
				}
			}
			w.Write([]byte("OK"))
		})

		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)
	})

	t.Run("does not error on empty session", func(t *testing.T) {
		s := server.New(
			server.WithSession("secret", "test"),
		)

		s.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
			err := session.Clear(r, w)
			if err != nil {
				t.Errorf("Clear should not fail on empty session: %v", err)
			}
			w.Write([]byte("OK"))
		})

		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)
	})

	t.Run("returns error when session not in context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()

		err := session.Clear(req, res)
		if !errors.Is(err, session.ErrSessionNotFound) {
			t.Errorf("Expected ErrSessionNotFound, got %v", err)
		}
	})
}

func TestSave(t *testing.T) {
	t.Run("saves session successfully", func(t *testing.T) {
		s := server.New(
			server.WithSession("secret", "test"),
		)

		s.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
			session.Set(r, w, "key", "value")

			err := session.Save(r, w)
			if err != nil {
				t.Errorf("Save failed: %v", err)
			}
			w.Write([]byte("OK"))
		})

		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)

		// Verify cookie was set (session was saved)
		cookies := res.Header()["Set-Cookie"]
		if len(cookies) == 0 {
			t.Error("Expected Set-Cookie header after Save")
		}
	})

	t.Run("returns error when session not in context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()

		err := session.Save(req, res)
		if !errors.Is(err, session.ErrSessionNotFound) {
			t.Errorf("Expected ErrSessionNotFound, got %v", err)
		}
	})
}

func TestValuesPersistence(t *testing.T) {
	t.Run("values persist across requests", func(t *testing.T) {
		var savedCtx context.Context

		s := server.New(
			server.WithSession("secret", "test"),
		)

		s.Use(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				h.ServeHTTP(w, r)
				savedCtx = context.WithoutCancel(r.Context())
			})
		})

		s.HandleFunc("GET /set/{$}", func(w http.ResponseWriter, r *http.Request) {
			session.Set(r, w, "persistent", "data")
			w.Write([]byte("OK"))
		})

		s.HandleFunc("GET /get/{$}", func(w http.ResponseWriter, r *http.Request) {
			val, exists, _ := session.Get(r, "persistent")
			if !exists || val != "data" {
				t.Errorf("Expected 'data', got '%v' (exists: %v)", val, exists)
			}
			w.Write([]byte("OK"))
		})

		// Set value in first request
		req := httptest.NewRequest("GET", "/set/", nil)
		res := httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)

		// Get value in second request using saved context
		req = httptest.NewRequest("GET", "/get/", nil)
		req = req.WithContext(savedCtx)
		res = httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)
	})

	t.Run("delete persists across requests", func(t *testing.T) {
		var savedCtx context.Context

		s := server.New(
			server.WithSession("secret", "test"),
		)

		s.Use(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				h.ServeHTTP(w, r)
				savedCtx = context.WithoutCancel(r.Context())
			})
		})

		s.HandleFunc("GET /set/{$}", func(w http.ResponseWriter, r *http.Request) {
			session.Set(r, w, "to_delete", "value")
			w.Write([]byte("OK"))
		})

		s.HandleFunc("GET /delete/{$}", func(w http.ResponseWriter, r *http.Request) {
			session.Delete(r, w, "to_delete")
			w.Write([]byte("OK"))
		})

		s.HandleFunc("GET /check/{$}", func(w http.ResponseWriter, r *http.Request) {
			_, exists, _ := session.Get(r, "to_delete")
			if exists {
				t.Error("Expected key to not exist after delete")
			}
			w.Write([]byte("OK"))
		})

		// Set value
		req := httptest.NewRequest("GET", "/set/", nil)
		res := httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)

		// Delete value
		req = httptest.NewRequest("GET", "/delete/", nil)
		req = req.WithContext(savedCtx)
		res = httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)

		// Check it's gone
		req = httptest.NewRequest("GET", "/check/", nil)
		req = req.WithContext(savedCtx)
		res = httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)
	})

	t.Run("clear persists across requests", func(t *testing.T) {
		var savedCtx context.Context

		s := server.New(
			server.WithSession("secret", "test"),
		)

		s.Use(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				h.ServeHTTP(w, r)
				savedCtx = context.WithoutCancel(r.Context())
			})
		})

		s.HandleFunc("GET /set/{$}", func(w http.ResponseWriter, r *http.Request) {
			session.Set(r, w, "key1", "value1")
			session.Set(r, w, "key2", "value2")
			w.Write([]byte("OK"))
		})

		s.HandleFunc("GET /clear/{$}", func(w http.ResponseWriter, r *http.Request) {
			session.Clear(r, w)
			w.Write([]byte("OK"))
		})

		s.HandleFunc("GET /check/{$}", func(w http.ResponseWriter, r *http.Request) {
			_, exists1, _ := session.Get(r, "key1")
			_, exists2, _ := session.Get(r, "key2")
			if exists1 || exists2 {
				t.Error("Expected all keys to not exist after clear")
			}
			w.Write([]byte("OK"))
		})

		// Set values
		req := httptest.NewRequest("GET", "/set/", nil)
		res := httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)

		// Clear all
		req = httptest.NewRequest("GET", "/clear/", nil)
		req = req.WithContext(savedCtx)
		res = httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)

		// Check they're gone
		req = httptest.NewRequest("GET", "/check/", nil)
		req = req.WithContext(savedCtx)
		res = httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)
	})
}
