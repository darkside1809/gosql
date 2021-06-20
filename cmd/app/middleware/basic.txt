package middleware

import (
	"encoding/base64"
	"net/http"
	"strings"
	"errors"

)

//Basic...
func Basic(checkAuth func(string, string) bool) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			// extract username and password
			login, password, err := getLogPass(*r)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized),http.StatusUnauthorized)
				return
			}

			if !checkAuth(login, password) {
				http.Error(w, http.StatusText(http.StatusUnauthorized),http.StatusUnauthorized)
				return
			}

			handler.ServeHTTP(w, r)
		})
	}
}

//func that extracts data from the request and returns the login and password
func getLogPass(r http.Request) (string, string, error) {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
			if len(s) != 2 {
				return "","",errors.New("invalid auth method")
			}
			b, err := base64.StdEncoding.DecodeString(s[1])
			if err != nil {
				return "","",errors.New("invalid auth method")
			}
			pair := strings.SplitN(string(b), ":", 2)
			if len(pair) != 2 {
				return "","",errors.New("invalid auth method")
			}
			return pair[0],pair[1],nil
}
