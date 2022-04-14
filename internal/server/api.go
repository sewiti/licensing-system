package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

const maxRequestSize = 256 * 1024 // 256 KiB

type apiHandler func(r *http.Request) *apiResponse

func withAPI(h apiHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPatch, http.MethodPut:
			contentType := r.Header.Get("Content-Type")
			i := strings.IndexRune(contentType, ';')
			if i >= 0 {
				contentType = contentType[:i]
			}
			if contentType != "application/json" {
				responseJsonMsg(http.StatusUnsupportedMediaType,
					"unsupported content type, expected application/json").
					Write(w)
				return
			}
		}
		h(r).Write(w)
	})
}

func jsonDecodeLim(r io.Reader, data interface{}) error {
	limited := io.LimitReader(r, maxRequestSize)
	return json.NewDecoder(limited).Decode(data)
}

func readAllLim(r io.Reader) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r, maxRequestSize))
}
