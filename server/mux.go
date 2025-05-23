// Package server provides HTTP server functionality with routing capabilities,
// middleware support, and session management built on top of standard Go net/http.
// It offers a simple API for building web applications with clean routing,
// middleware chains, and error handling.
package server

import (
	"fmt"
	"net/http"
)

// defaultCatchAllHandler to log and return a 404 for all routes except the root route.
var defaultCatchAllHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		return
	}

	Error(w, fmt.Errorf("404 page not found"), http.StatusNotFound)
})

// mux is the root server that contains a router group with a common prefix and middleware.
// It also has a host and port configuration and serves as the main entry point
// for handling HTTP requests.
type mux struct {
	*router

	host string
	port string
}

// New creates a new server with the given options and default middleware.
func New(options ...Option) *mux {
	ss := &mux{
		router: &router{
			prefix:     "",
			mux:        http.NewServeMux(),
			middleware: baseMiddleware,
		},

		host: "0.0.0.0",
		port: "3000",
	}

	for _, option := range options {
		option(ss)
	}

	return ss
}

func (s *mux) Router() Router {
	return s.router
}

func (s *mux) Handler() http.Handler {
	// if no catch-all or root route has been set
	// we use the default one
	if !s.rootSet {
		s.Handle("/", defaultCatchAllHandler)
	}

	return s
}

func (s *mux) Addr() string {
	return s.host + ":" + s.port
}
