package middlewares

import (
	"net/http"

	"github.com/juliotorresmoreno/zemona/helpers"
)

func Cors(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			helpers.Cors(w, r)
			handler.ServeHTTP(w, r)
			return
		}
		if r.Method == "OPTIONS" {
			helpers.Cors(w, r)
			w.WriteHeader(http.StatusOK)
			return
		}
		handler.ServeHTTP(w, r)
	})
}
