package server

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type corsHandler struct {
	headers []string
	methods []string
}

func (h corsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(h.headers, ","))
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(h.methods, ","))
	w.WriteHeader(http.StatusNoContent)
}

func corsMethodMiddleware(r *mux.Router) func(http.Handler) http.Handler {
	mid := mux.CORSMethodMiddleware(r)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				h = mid(h)
			}
			h.ServeHTTP(w, r)
		})
	}
}

func corsOriginMiddleware(origins []string) func(http.Handler) http.Handler {
	allowed := func(origin string) (string, bool) {
		for _, v := range origins {
			switch v {
			case origin, "*":
				return v, true
			}
		}
		return origin, false
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			origin, ok := allowed(origin)
			if ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				if len(origins) > 1 {
					w.Header().Set("Vary", "Origin")
				}
			}
			h.ServeHTTP(w, r)
		})
	}
}
