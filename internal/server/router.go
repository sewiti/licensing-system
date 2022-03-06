package server

import (
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sewiti/licensing-system/internal/core"
)

func NewRouter(c *core.Core, resourceApiCors, licensingCors bool, allowedOrigins []string) *mux.Router {
	withAPIAuthorized := func(h apiAuthHandler) http.Handler {
		return withAPI(withAPIAuth(c, withAuthorized(c, h)))
	}

	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	api.Use(func(h http.Handler) http.Handler {
		return handlers.ContentTypeHandler(h, "application/json") // Applies to POST, PUT and PATCH
	})

	// Licensing API
	api.Path("/license-sessions").Methods(http.MethodPost).Handler(withAPI(licCreateLicenseSession(c)))
	api.Path("/license-sessions/{CLIENT_SESSION_ID:[A-Za-z0-9_-]{43}=}").Methods(http.MethodPatch).Handler(withAPI(licUpdateLicenseSession(c)))
	api.Path("/license-sessions/{CLIENT_SESSION_ID:[A-Za-z0-9_-]{43}=}").Methods(http.MethodDelete).Handler(withAPI(licDeleteLicenseSession(c)))

	if licensingCors {
		api.PathPrefix("/license-sessions").Subrouter().Use(
			handlers.CORS(
				handlers.AllowedOrigins(allowedOrigins),
				handlers.AllowedHeaders([]string{"Content-Type"}),
				handlers.AllowedMethods([]string{http.MethodPost, http.MethodPatch, http.MethodDelete, http.MethodOptions}),
			),
		)
	}

	// Resource API
	api.Path("/license-issuers").Methods(http.MethodPost).Handler(withAPIAuthorized(createLicenseIssuer(c)))
	api.Path("/license-issuers").Methods(http.MethodGet).Handler(withAPIAuthorized(getAllLicenseIssuers(c)))
	api.Path("/license-issuers/{LICENSE_ISSUER_ID:[0-9]+}").Methods(http.MethodGet).Handler(withAPIAuthorized(getLicenseIssuer(c)))
	api.Path("/license-issuers/{LICENSE_ISSUER_ID:[0-9]+}").Methods(http.MethodPatch).Handler(withAPIAuthorized(updateLicenseIssuer(c)))
	api.Path("/license-issuers/{LICENSE_ISSUER_ID:[0-9]+}").Methods(http.MethodDelete).Handler(withAPIAuthorized(deleteLicenseIssuer(c)))

	apili := api.PathPrefix("/license-issuers/{LICENSE_ISSUER_ID:[0-9]+}").Subrouter()
	apili.Path("/licenses").Methods(http.MethodPost).Handler(withAPIAuthorized(createLicense(c)))
	apili.Path("/licenses").Methods(http.MethodGet).Handler(withAPIAuthorized(getAllLicenses(c)))
	apili.Path("/licenses/{LICENSE_ID:[A-Za-z0-9_-]{43}=}").Methods(http.MethodGet).Handler(withAPIAuthorized(getLicense(c)))
	apili.Path("/licenses/{LICENSE_ID:[A-Za-z0-9_-]{43}=}").Methods(http.MethodPatch).Handler(withAPIAuthorized(updateLicense(c)))
	apili.Path("/licenses/{LICENSE_ID:[A-Za-z0-9_-]{43}=}").Methods(http.MethodDelete).Handler(withAPIAuthorized(deleteLicense(c)))

	apilil := apili.PathPrefix("/licenses/{LICENSE_ID:[A-Za-z0-9_-]{43}=}").Subrouter()
	apilil.Path("/sessions").Methods(http.MethodGet).Handler(withAPIAuthorized(getAllLicenseSessions(c)))
	apilil.Path("/sessions/{CLIENT_SESSION_ID:[A-Za-z0-9_-]{43}=}").Methods(http.MethodGet).Handler(withAPIAuthorized(getLicenseSession(c)))
	apilil.Path("/sessions/{CLIENT_SESSION_ID:[A-Za-z0-9_-]{43}=}").Methods(http.MethodDelete).Handler(withAPIAuthorized(deleteLicenseSession(c)))

	if resourceApiCors {
		api.PathPrefix("/license-issuers").Subrouter().Use(
			handlers.CORS(
				handlers.AllowedOrigins(allowedOrigins),
				handlers.AllowedHeaders([]string{"Authorization", "Content-Type"}),
				handlers.AllowedMethods([]string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete, http.MethodOptions}),
			),
		)
	}

	// Auth API
	api.Path("/login").Methods(http.MethodPost).Handler(withAPI(createToken(c)))
	api.Path("/change-password/{LICENSE_ISSUER_ID:[0-9]+}").Methods(http.MethodPatch).Handler(withAPIAuthorized(updatePassword(c)))

	if resourceApiCors {
		api.Path("/login").Subrouter().Use(handlers.CORS(
			handlers.AllowedOrigins(allowedOrigins),
			handlers.AllowedHeaders([]string{"Content-Type"}),
			handlers.AllowedMethods([]string{http.MethodPost, http.MethodOptions}),
		))
		api.Path("/change-password/{LICENSE_ISSUER_ID:[0-9]+}").Subrouter().Use(handlers.CORS(
			handlers.AllowedOrigins(allowedOrigins),
			handlers.AllowedHeaders([]string{"Authorization", "Content-Type"}),
			handlers.AllowedMethods([]string{http.MethodPatch, http.MethodOptions}),
		))
	}

	// Single page app
	// TODO

	return r
}

func pathVarID(str string) (*[32]byte, error) {
	if str == "" {
		return nil, errors.New("missing var")
	}
	bs, err := base64.URLEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}
	return (*[32]byte)(bs), nil
}
