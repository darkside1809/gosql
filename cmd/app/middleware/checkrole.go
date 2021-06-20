package middleware

import (
	"context"
	"net/http"
)

type HasAnyRoleFunc func(ctx context.Context, roles ...string) bool

func CheckRole(hasAnyRoleFunc HasAnyRoleFunc, roles ...string) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !hasAnyRoleFunc(r.Context(), roles ...) {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			handler.ServeHTTP(w, r)
		})
	}
}