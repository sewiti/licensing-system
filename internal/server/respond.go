package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/apex/log"
)

type jsonResponse struct {
	statusCode int
	data       interface{}
}

func (r *jsonResponse) respond(w http.ResponseWriter) {
	if r.statusCode == http.StatusNoContent {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(r.statusCode)

	err := json.NewEncoder(w).Encode(r.data)
	if err != nil {
		log.WithError(err).
			WithField("data", r.data).
			Error("encoding json response")
	}
}

type messageResponse struct {
	Message string `json:"message"`
}

var errNotFound = &jsonResponse{
	statusCode: http.StatusNotFound,
	data:       messageResponse{Message: "not found"},
}

var errInternalServerError = &jsonResponse{
	statusCode: http.StatusInternalServerError,
	data:       messageResponse{Message: "internal server error"},
}

var respondNoContent = &jsonResponse{
	statusCode: http.StatusNoContent,
}

func respondJson(statusCode int, data interface{}) *jsonResponse {
	return &jsonResponse{
		statusCode: statusCode,
		data:       data,
	}
}

func respondMessage(statusCode int, a ...interface{}) *jsonResponse {
	return respondJson(statusCode,
		messageResponse{Message: fmt.Sprint(a...)},
	)
}

func respondMessagef(statusCode int, format string, a ...interface{}) *jsonResponse {
	return respondJson(statusCode,
		messageResponse{Message: fmt.Sprintf(format, a...)},
	)
}

func respondBadRequest(a ...interface{}) *jsonResponse {
	return respondMessage(http.StatusBadRequest, a...)
}

func respondBadRequestf(format string, a ...interface{}) *jsonResponse {
	return respondMessagef(http.StatusBadRequest, format, a...)
}
