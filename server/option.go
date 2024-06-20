package server

import (
	"net/http"

	"github.com/leapkit/core/session"
)

// Options for the server
type Option func(*mux)

// WithHost allows to specify the host to run the server at
// if not specified it defaults to 0.0.0.0
func WithHost(host string) Option {
	return func(s *mux) {
		s.host = host
	}
}

// WithPort allows to specify the port to run the server at
// when not specified it defaults to 3000
func WithPort(port string) Option {
	return func(s *mux) {
		s.port = port
	}
}

func WithSession(secret, name string) Option {
	return func(m *mux) {
		m.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				session := session.New(secret, name)
				session.Register(w, r)

				next.ServeHTTP(w, r)
			})
		})
	}
}
