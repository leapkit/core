package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Instance of the server, it contains a router and its basic options,
// this instance is used to apply options to the server.
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

func (r *Instance) Folder(path string, dir fs.FS) {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	// mount the folder at the given path with the given prefix
	r.Handle(path+"*", http.StripPrefix(path, http.FileServer(http.FS(dir))))
}

// New sets up and returns a new HTTP server with routes mounted
// for each of the different features in this application. It also
// sets up the default middleware for the server.
func New(name string, options ...Option) *Instance {
	r := &Instance{
		Mux:  chi.NewRouter(),
		name: name,
		host: "0.0.0.0", //default host
		port: "3000",    //default port
	}

	r.Use(setValuer)

	for _, v := range options {
		v(r)
	}

	return r
}
