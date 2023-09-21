package render

import (
	"context"
	"net/http"
)

// InCtx puts the render engine in the context
// so the handlers can use it, it also sets a few
// other values that are useful for the handlers.
func InCtx(engine *Engine) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			px := engine.HTML(w)

			px.Set("request", r)
			px.Set("currentURL", r.URL.String())

			r = r.WithContext(context.WithValue(r.Context(), "renderer", px))
			next.ServeHTTP(w, r)
		})
	}
}
