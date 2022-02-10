package server

import (
	"encoding/json"
	"io"
	"net/http"
)

type apiHandler func(r *http.Request) *jsonResponse

func withAPI(h apiHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h(r).respond(w)
	})
}

func jsonDecodeLim(r io.Reader, data interface{}) error {
	const maxRequest = 256 * 1024 // 256 KiB
	limited := io.LimitReader(r, maxRequest)
	return json.NewDecoder(limited).Decode(data)
}
