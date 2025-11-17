package server

import (
	"context"
	"net/http"
	"sync"
)

const (
	valuerContextKey = "valuer"
)

// Valuer interface to store and retrieve values from the request context
// these values can be used by other components such as the render engine
// or the assets manager.
type Valuer interface {
	Values() map[string]any
	Value(string) any
	Set(string, any)
}

// ValuerFromContext retrieves the Valuer from the context.
func ValuerFromContext(ctx context.Context) Valuer {
	vlr, ok := ctx.Value(valuerContextKey).(Valuer)
	if !ok {
		return nil
	}

	return vlr
}

// values is a map where we can store values for the request context
// these values will then be available for other components such as
// the render engine.
type valuer struct {
	data map[string]any
	moot sync.Mutex
}

// Value returns the value for the key specified.
func (v *valuer) Value(key string) any {
	return v.data[key]
}

// Values returns the values stored in the valuer.
func (v *valuer) Values() map[string]any {
	return v.data
}

// Set sets the value for the key specified.
func (v *valuer) Set(key string, value any) {
	v.moot.Lock()
	defer v.moot.Unlock()

	v.data[key] = value
}

// For each of the requests we want to have a valuer instance so
// that we can store values in the context that can be used by
// other components.
func setValuer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vlr := &valuer{
			data: map[string]any{
				// Adding base values that are useful for the handlers.
				"request":    r,
				"currentURL": r.URL.String(),
			},
		}

		r = r.WithContext(context.WithValue(r.Context(), valuerContextKey, vlr))
		next.ServeHTTP(w, r)
	})
}
