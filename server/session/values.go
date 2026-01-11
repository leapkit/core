package session

import (
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
)

// ErrSessionNotFound is returned when the session is not found in the request context.
var ErrSessionNotFound = errors.New("session not found in request context")

// Get retrieves a value from the session by key.
// Returns the value and true if found, or nil and false if not found.
// Returns an error if the session is not in the request context.
func Get(r *http.Request, key string) (any, bool, error) {
	session, ok := r.Context().Value(ctxKey).(*sessions.Session)
	if !ok || session == nil {
		return nil, false, ErrSessionNotFound
	}

	val, exists := session.Values[key]
	return val, exists, nil
}

// Set stores a value in the session.
// Returns an error if the session is not in the request context.
func Set(r *http.Request, w http.ResponseWriter, key string, value any) error {
	session, ok := r.Context().Value(ctxKey).(*sessions.Session)
	if !ok || session == nil {
		return ErrSessionNotFound
	}

	session.Values[key] = value
	return nil
}

// Delete removes a value from the session by key.
// Returns an error if the session is not in the request context.
func Delete(r *http.Request, w http.ResponseWriter, key string) error {
	session, ok := r.Context().Value(ctxKey).(*sessions.Session)
	if !ok || session == nil {
		return ErrSessionNotFound
	}

	delete(session.Values, key)
	return nil
}

// Clear removes all values from the session.
// Returns an error if the session is not in the request context.
func Clear(r *http.Request, w http.ResponseWriter) error {
	session, ok := r.Context().Value(ctxKey).(*sessions.Session)
	if !ok || session == nil {
		return ErrSessionNotFound
	}

	for k := range session.Values {
		delete(session.Values, k)
	}

	return nil
}

// Save persists the current session state.
// Note: Sessions are automatically saved when the response is written,
// so this is only needed in special cases like WebSocket upgrades or
// long-running handlers where you want to persist changes early.
func Save(r *http.Request, w http.ResponseWriter) error {
	session, ok := r.Context().Value(ctxKey).(*sessions.Session)
	if !ok || session == nil {
		return ErrSessionNotFound
	}

	return session.Save(r, w)
}
