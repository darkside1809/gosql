package middleware

import (
	"log"
	"net/http"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// Basic makes basic authintication by provided username and password
func Basic(auth func(login, pass string) bool) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			username, password, ok := request.BasicAuth()
			if !ok {
				log.Print("Cant parse username and password")
				http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			if !auth(username, password) {
				http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			handler.ServeHTTP(writer, request)
		})
	}
}

// Logger print handler's paths and methods which were called
func Logger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Start: %s %s", r.Method, r.URL.Path)
		handler.ServeHTTP(w, r)
		log.Printf("Finish: %s %s", r.Method, r.URL.Path)
	})
}

// CheckHeader checks definite handler's request on header appropriation
func CheckHeader(header string, value string) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if value != r.Header.Get(header) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			handler.ServeHTTP(w, r)
		})
	}
}