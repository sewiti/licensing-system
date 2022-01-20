package web

import (
	"github.com/gorilla/mux"
)

func NewRouter(rt *Runtime) *mux.Router {
	r := mux.NewRouter()

	return r
}
