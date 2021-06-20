package middleware

import (
	"net/http"
	"errors"
	"context"
)

type contextKey struct {
	name string
}
type IDFunc func(ctx context.Context, token string) (int64, error)

var ErrNoAuthentication = errors.New("no authentication")
var authenticationContextKey = &contextKey{"authentication context"}

func (c *contextKey) String() string{
	return c.name
}

// Authenticate by customer token
func Authenticate(idFunc IDFunc) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")

			id, err := idFunc(r.Context(), token)
			if err == nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), authenticationContextKey, id)
			r = r.WithContext(ctx)
			handler.ServeHTTP(w, r)
		}) 
	}
}

func Authentication(ctx context.Context) (int64, error) {
	if value, ok := ctx.Value(authenticationContextKey).(int64); ok {
		return value, nil
	}
	return 0, ErrNoAuthentication
}