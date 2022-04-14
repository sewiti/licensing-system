package server

import (
	"net/url"
	"strconv"
)

type pageResponse struct {
	Total int         `json:"total"`
	Data  interface{} `json:"data"`
}

func paginationVars(query url.Values) (limit, offset int, ok bool, err error) {
	if !query.Has("limit") && !query.Has("offset") {
		return 0, 0, false, nil
	}

	limit, err = strconv.Atoi(query.Get("limit"))
	if err != nil {
		return 0, 0, true, err
	}
	offset, err = strconv.Atoi(query.Get("offset"))
	if err != nil {
		return 0, 0, true, err
	}
	return limit, offset, true, err
}
