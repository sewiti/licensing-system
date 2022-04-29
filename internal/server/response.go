package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/apex/log"
)

type apiResponse struct {
	statusCode int
	json       bool
	body       []byte
}

func (res *apiResponse) Write(w http.ResponseWriter) error {
	if res == nil || res.statusCode == http.StatusNoContent {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	if res.json {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	w.WriteHeader(res.statusCode)
	_, err := w.Write(res.body)
	return err
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

func spaHandler(public fs.FS, root string) http.Handler {
	open := func(path string) (fs.File, error) {
		f, err := public.Open(filepath.Join(root, path))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return public.Open(filepath.Join(root, "index.html"))
			}
			return nil, err
		}
		fi, err := f.Stat()
		if err != nil {
			_ = f.Close()
			return nil, err
		}
		if fi.IsDir() {
			_ = f.Close()
			return public.Open(filepath.Join(root, "index.html"))
		}
		return f, nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := path.Clean(r.URL.Path)
		f, err := open(p)
		if err != nil {
			log.WithError(err).WithField("path", p).Error("serving spa, opening file")
			responseInternalServerError().Write(w)
			return
		}
		defer f.Close()
		switch {
		case strings.HasSuffix(p, ".js"):
			w.Header().Set("Content-Type", "text/javascript")
		case strings.HasSuffix(p, ".css"):
			w.Header().Set("Content-Type", "text/css")
		}
		_, err = io.Copy(w, f)
		if err != nil {
			log.WithError(err).WithField("path", p).Error("serving spa")
		}
	})
}
