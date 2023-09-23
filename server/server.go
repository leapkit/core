package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Instance struct {
	*chi.Mux

	name string
	host string
	port string
}

func (r *Instance) Start() error {
	host := fmt.Sprintf("%v:%v", r.host, r.port)
	fmt.Printf("[info] Starting %v server on port %v\n", r.name, host)

	return http.ListenAndServe(host, r)
}

// New sets up and returns a new HTTP server with routes mounted
// for each of the different features in this application.
func New(name string, options ...Option) *Instance {
	r := &Instance{
		Mux:  chi.NewRouter(),
		name: name,
		host: "0.0.0.0",
		port: "3000",
	}

	for _, v := range options {
		v(r)
	}

	return r
}
