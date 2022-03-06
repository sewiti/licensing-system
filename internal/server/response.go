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

type messageResponse struct {
	Message string `json:"message"`
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
		messageResponse{
			Message: fmt.Sprint(a...),
		})
}

func responseJsonMsgf(statusCode int, format string, a ...interface{}) *apiResponse {
	return responseJson(statusCode,
		messageResponse{
			Message: fmt.Sprintf(format, a...),
		})
}

func responseNoContent() *apiResponse {
	return &apiResponse{
		statusCode: http.StatusNoContent,
	}
}

func responseBadRequest(a ...interface{}) *apiResponse {
	if len(a) == 0 {
		return responseJsonMsg(http.StatusBadRequest, "400 Bad Request")
	}
	return responseJsonMsg(http.StatusBadRequest, a...)
}

func responseBadRequestf(format string, a ...interface{}) *apiResponse {
	return responseJsonMsgf(http.StatusBadRequest, format, a...)
}

func responseUnauthorized() *apiResponse {
	return responseJsonMsg(http.StatusUnauthorized, "401 Unauthorized")
}

func responseForbidden(a ...interface{}) *apiResponse {
	if len(a) == 0 {
		return responseJsonMsg(http.StatusForbidden, "403 Forbidden")
	}
	return responseJsonMsg(http.StatusForbidden, a...)
}

// func responseForbiddenf(format string, a ...interface{}) *apiResponse {
// 	return responseJsonMsgf(http.StatusForbidden, format, a...)
// }

func responseNotFound() *apiResponse {
	return responseJsonMsg(http.StatusNotFound, "404 Not Found")
}

func responseConflict(a ...interface{}) *apiResponse {
	if len(a) == 0 {
		return responseJsonMsg(http.StatusConflict, "409 Conflict")
	}
	return responseJsonMsg(http.StatusConflict, a...)
}

// func responseConflictf(format string, a ...interface{}) *apiResponse {
// 	return responseJsonMsgf(http.StatusConflict, format, a...)
// }

func responseInternalServerError() *apiResponse {
	return &apiResponse{
		statusCode: http.StatusInternalServerError,
		body:       []byte("500 Internal Server Error"),
	}
}
