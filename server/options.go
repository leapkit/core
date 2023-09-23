package server

import (
	"context"
	"net/http"
)

// Allows to specify options to the server.
type Option func(*Instance)

// Allows to specify a key/value that should be set on each
// request context. This is useful for services that could be
// used by the handlers.
func WithCtxVal(key string, value interface{}) Option {
	return func(s *Instance) {
		s.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r = r.WithContext(context.WithValue(r.Context(), key, value))
				next.ServeHTTP(w, r)
			})
		})
	}
}

func WithRoutesFn(routesFn func(*Instance)) Option {
	return func(s *Instance) {
		routesFn(s)
	}
}

func WithPort(port int) Option {
	return func(s *Instance) {
		s.port = port
	}
}
