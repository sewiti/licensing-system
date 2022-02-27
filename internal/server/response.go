package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/apex/log"
)

type apiResponse struct {
	statusCode int
	json       bool
	body       []byte
}

func responseJson(statusCode int, data interface{}) *apiResponse {
	bs, err := json.Marshal(data)
	if err != nil {
		log.WithError(err).Error("marshaling json response")
		return responseInternalServerError()
	}
	return &apiResponse{
		statusCode: statusCode,
		json:       true,
		body:       bs,
	}
}

func responseJsonMsg(statusCode int, a ...interface{}) *apiResponse {
	return responseJson(statusCode,
		struct {
			Msg string `json:"message"`
		}{
			Msg: fmt.Sprint(a...),
		},
	)
}

func responseJsonMsgf(statusCode int, format string, a ...interface{}) *apiResponse {
	return responseJson(statusCode,
		struct {
			Msg string `json:"message"`
		}{
			Msg: fmt.Sprintf(format, a...),
		},
	)
}

func responseNoContent() *apiResponse {
	return &apiResponse{
		statusCode: http.StatusNoContent,
	}
}

func responseBadRequest(a ...interface{}) *apiResponse {
	return responseJsonMsg(http.StatusBadRequest, a...)
}

func responseBadRequestf(format string, a ...interface{}) *apiResponse {
	return responseJsonMsgf(http.StatusBadRequest, format, a...)
}

func responseForbidden(a ...interface{}) *apiResponse {
	return responseJsonMsg(http.StatusForbidden, a...)
}

func responseNotFound() *apiResponse {
	return &apiResponse{
		statusCode: http.StatusNotFound,
		body:       []byte("404 Not Found"),
	}
}

func responseConflict(a ...interface{}) *apiResponse {
	return responseJsonMsg(http.StatusConflict, a...)
}

func responseInternalServerError() *apiResponse {
	return &apiResponse{
		statusCode: http.StatusInternalServerError,
		body:       []byte("500 Internal Server Error"),
	}
}
