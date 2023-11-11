package render

import (
	"context"
	"net/http"
)

// InCtx puts the render engine in the context
// so the handlers can use it, it also sets a few
// other values that are useful for the handlers.
var InCtx = Middleware

// Middleware puts the render engine in the context
// so the handlers can use it, it also sets a few
// other values that are useful for the handlers.
func Middleware(engine *Engine) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "renderer", engine.HTML(w))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
