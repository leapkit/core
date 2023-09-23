package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Instance struct {
	*chi.Mux

	name string
	port string
}

func (r *Instance) Start() error {
	fmt.Printf("[info] Starting %v server on port %v\n", r.name, r.port)
	return http.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", r.port), r)
}

// New sets up and returns a new HTTP server with routes mounted
// for each of the different features in this application.
func New(name string, options ...Option) *Instance {
	r := &Instance{
		Mux:  chi.NewRouter(),
		name: name,
		port: "3000",
	}

	for _, v := range options {
		v(r)
	}

	return r
}
