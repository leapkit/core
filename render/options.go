package render

import "html/template"

type Option func(*Engine)

func WithHelpers(hps template.FuncMap) Option {
	return func(e *Engine) {
		e.helpers = hps
	}
}
