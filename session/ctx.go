package session

import (
	"context"
	"encoding/gob"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/sessions"
)

var ctxKey = "session"

func init() {
	gob.Register(uuid.UUID{})
}

// InCtx is a middleware that injects the session into the request context
// and also takes care of saving the session when the response is written
// to the client by wrapping the response writer.
func InCtx(store *sessions.CookieStore, name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := store.Get(r, name)
			r = r.WithContext(context.WithValue(r.Context(), ctxKey, session))
			w = &saver{
				w:     w,
				req:   r,
				store: session,
			}

			next.ServeHTTP(w, r)
		})
	}
}

// FromCtx returns the session from the context.
func FromCtx(ctx context.Context) *sessions.Session {
	return ctx.Value(ctxKey).(*sessions.Session)
}
