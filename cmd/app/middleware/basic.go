package middleware

import (
	"log"
	"net/http"
	"errors"
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
)

var ErrNoAuthentication = errors.New("no authentication") 
var AuthenticationCtxKey = &contextKey{"authentication context"}
type Service struct {
	pool *pgxpool.Pool
}
type contextKey struct {
	name string
}
type IDFunc func(ctx context.Context, token string) (int64, error)

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// Basic makes basic authintication by provided username and password
func Basic(auth func(login, pass string) bool) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			username, password, ok := request.BasicAuth()
			if !ok {
				log.Print("Cant parse username or password")
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

// Authenticate by customer token
func Authenticate(IDFunc IDFunc) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			token := request.Header.Get("Authorization")
			log.Print(token)
			id, err := IDFunc(request.Context(), token)
			if err != nil {
				http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			ctx := context.WithValue(request.Context(), AuthenticationCtxKey, id)
			request = request.WithContext(ctx)
			handler.ServeHTTP(writer, request)
		}) 
	}
}

func Authentication(ctx context.Context) (int64, error) {
	if value, ok := ctx.Value(AuthenticationCtxKey).(int64); ok {
		return value, nil
	}
	return 0, ErrNoAuthentication
}