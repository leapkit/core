package render

import (
	"html/template"
	"io/fs"
)

type Engine struct {
	templates fs.FS
	helpers   template.FuncMap
	values    map[string]any
}

func (e *Engine) Set(key string, value any) {
	e.values[key] = value
}

func (e *Engine) SetHelper(key string, value any) {
	e.helpers[key] = value
}

func NewEngine(fs fs.FS, options ...Option) *Engine {
	e := &Engine{
		templates: fs,

		values:  make(map[string]any),
		helpers: make(template.FuncMap),
	}

	for _, option := range options {
		option(e)
	}

	return e
}
