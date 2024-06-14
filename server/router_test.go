package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leapkit/core/server"
)

func TestRouter(t *testing.T) {

	s := server.New()

	s.Group("/", func(r server.Router) {
		r.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello, World!"))
		})

		r.Group("/api/", func(r server.Router) {
			r.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("This is the API!"))
			})

			r.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("API documentation!"))
			})

			r.Group("/v1/", func(r server.Router) {
				r.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("Welcome to the API v1!"))
				})

				r.Group("/users/", func(r server.Router) {
					r.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
						w.Write([]byte("Users list!"))
					})

					r.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
						w.Write([]byte("Hello users!"))
					})
				})
			})
		})
	})

	testCases := []struct {
		method string
		route  string
		body   string
		code   int
	}{
		{"GET", "/", "Hello, World!", http.StatusOK},
		{"GET", "/api/v1/users/hello", "Hello users!", http.StatusOK},
		{"GET", "/api/v1/users/", "Users list!", http.StatusOK},
		{"GET", "/api/v1/", "Welcome to the API v1!", http.StatusOK},
		{"GET", "/api/", "This is the API!", http.StatusOK},
		{"GET", "/api/docs", "API documentation!", http.StatusOK},
	}

	for _, tt := range testCases {
		t.Run(tt.route, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.route, nil)
			res := httptest.NewRecorder()
			s.Handler().ServeHTTP(res, req)

			if res.Code != tt.code {
				t.Errorf("Expected status code %d, got %d", tt.code, res.Code)
			}

			if res.Body.String() != tt.body {
				t.Errorf("Expected body %s, got %s", tt.body, res.Body.String())
			}
		})
	}
}

func Test_Router_Middleware(t *testing.T) {
	m := server.InCtxMiddleware("message", "Hello World!")

	t.Run("Setting middleware in router", func(t *testing.T) {
		s := server.New()
		s.Use(m)
		s.Group("/hello", func(rg server.Router) {
			rg.HandleFunc("GET /world", func(w http.ResponseWriter, r *http.Request) {
				message, ok := r.Context().Value("message").(string)
				if !ok {
					message = "message not found"
				}

				w.Write([]byte(message))
			})
		})

		req, _ := http.NewRequest("GET", "/hello/world", nil)
		res := httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, res.Code)
		}

		if res.Body.String() != "Hello World!" {
			t.Errorf("Expected body 'Hello World!', got %s", res.Body.String())
		}
	})

	t.Run("cleaning middlewares in router", func(t *testing.T) {
		s := server.New()
		s.Use(m)

		s.Group("/hello", func(rg server.Router) {
			rg.ClearMiddlewares()
			rg.HandleFunc("GET /world", func(w http.ResponseWriter, r *http.Request) {
				message, ok := r.Context().Value("message").(string)
				if !ok {
					message = "Message not found"
				}

				w.Write([]byte(message))
			})
		})

		req, _ := http.NewRequest("GET", "/hello/world", nil)
		res := httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, res.Code)
		}

		if res.Body.String() != "Message not found" {
			t.Errorf("Expected body 'Message not found', got %s", res.Body.String())
		}
	})

	t.Run("cleaning middlewares in router", func(t *testing.T) {
		s := server.New()
		s.Use(m)

		s.Group("/", func(r server.Router) {
			r.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
				message, ok := r.Context().Value("message").(string)
				if !ok {
					message = "Message not found"
				}

				w.Write([]byte(message))
			})

			r.Group("/hello/", func(r server.Router) {
				r.ClearMiddlewares()
				r.Use(server.InCtxMiddleware("message", "Hello people!"))
				r.HandleFunc("GET /people/{$}", func(w http.ResponseWriter, r *http.Request) {
					message, ok := r.Context().Value("message").(string)
					if !ok {
						message = "Message not found"
					}

					w.Write([]byte(message))
				})
			})
		})

		testCases := []struct {
			method string
			route  string
			body   string
			code   int
		}{
			{"GET", "/hello/people/", "Hello people!", http.StatusOK},
			{"GET", "/", "Hello World!", http.StatusOK},
		}

		for _, tt := range testCases {
			t.Run(tt.route, func(t *testing.T) {
				req, _ := http.NewRequest(tt.method, tt.route, nil)
				res := httptest.NewRecorder()
				s.Handler().ServeHTTP(res, req)

				if res.Code != tt.code {
					t.Errorf("Expected status code %d, got %d", tt.code, res.Code)
				}

				if res.Body.String() != tt.body {
					t.Errorf("Expected body %s, got %s", tt.body, res.Body.String())
				}
			})
		}
	})
}
