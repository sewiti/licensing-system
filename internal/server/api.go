package server

import (
	"encoding/json"
	"io"
	"net/http"
)

type apiHandler func(r *http.Request) *apiResponse

func withAPI(h apiHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := h(r)
		if res == nil || res.statusCode == http.StatusNoContent {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if res.json {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		} else {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		}
		w.WriteHeader(res.statusCode)
		w.Write(res.body)
	})
}

func jsonDecodeLim(r io.Reader, data interface{}) error {
	const maxRequest = 256 * 1024 // 256 KiB
	limited := io.LimitReader(r, maxRequest)
	return json.NewDecoder(limited).Decode(data)
}
