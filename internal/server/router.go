package server

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sewiti/licensing-system/internal/core"
)

func NewRouter(c *core.Core, useCors bool, allowedOrigins []string) *mux.Router {
	r := mux.NewRouter()

	apiMiddlewares := []mux.MiddlewareFunc{
		func(h http.Handler) http.Handler {
			return handlers.ContentTypeHandler(h, "application/json")
		},
	}
	if useCors {
		apiMiddlewares = append(apiMiddlewares,
			handlers.CORS(
				handlers.AllowedOrigins(allowedOrigins),
				handlers.AllowedHeaders([]string{"Authorization", "Content-Type"}),
				handlers.AllowedMethods([]string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete, http.MethodOptions}),
			))
	}

	api := r.PathPrefix("/api").Subrouter()
	api.Use(apiMiddlewares...)
	api.Path("/license-sessions").Methods(http.MethodPost).Handler(withAPI(createLicenseSession(c)))
	api.Path("/license-sessions/{CSID:[A-Za-z0-9_-]{43}=}").Methods(http.MethodPatch).Handler(withAPI(updateLicenseSession(c)))
	api.Path("/license-sessions/{CSID:[A-Za-z0-9_-]{43}=}").Methods(http.MethodDelete).Handler(withAPI(deleteLicenseSession(c)))

	return r
}
