package server

import (
	"net/http"
	"strings"
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
