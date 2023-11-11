package render

import (
	"html/template"
	"io/fs"
	"sync"
)

type Engine struct {
	templates     fs.FS
	defaultLayout string

	moot    sync.Mutex
	helpers template.FuncMap
	values  map[string]any
}

func (e *Engine) Set(key string, value any) {
	e.moot.Lock()
	defer e.moot.Unlock()

	e.values[key] = value
}

func (e *Engine) SetHelper(key string, value any) {
	e.moot.Lock()
	defer e.moot.Unlock()

	e.helpers[key] = value
}

func NewEngine(fs fs.FS, options ...Option) *Engine {
	e := &Engine{
		templates: fs,

		values:  make(map[string]any),
		helpers: make(template.FuncMap),

		defaultLayout: "app/layouts/application.html",
	}

	for _, option := range options {
		option(e)
	}

	return e
}
