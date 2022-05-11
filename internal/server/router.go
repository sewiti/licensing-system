package server

import (
	"embed"
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sewiti/licensing-system/internal/core"
)

//go:embed public/*
var publicDir embed.FS

func NewRouter(c *core.Core, resourceApiCors, licensingCors bool, allowedOrigins []string) *mux.Router {
	corsOriginMiddleware := corsOriginMiddleware(allowedOrigins)
	corsHandler := corsOriginMiddleware(corsHandler{
		headers: []string{"Authorization", "Content-Type"},
		methods: []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete, http.MethodOptions},
	})

	resourceHandler := func(r *mux.Router, path, method string, h http.Handler) {
		if resourceApiCors {
			h = corsOriginMiddleware(h)
			r.Path(path).Methods(http.MethodOptions).Handler(corsHandler)
		}
		r.Path(path).Methods(method).Handler(h)
	}
	licensingHandler := func(r *mux.Router, path, method string, h http.Handler) {
		if licensingCors {
			h = corsOriginMiddleware(h)
			r.Path(path).Methods(http.MethodOptions).Handler(corsHandler)
		}
		r.Path(path).Methods(method).Handler(h)
	}

	withAPIAuthorized := func(h apiAuthHandler) http.Handler {
		return withAPI(withAPIAuth(c, withAuthorized(c, h)))
	}

	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	// Licensing API
	licensingHandler(api, "/license-sessions", http.MethodPost, withAPI(licCreateLicenseSession(c)))
	licensingHandler(api, "/license-sessions/{CLIENT_SESSION_ID:[A-Za-z0-9_-]{43}=}", http.MethodPatch, withAPI(licUpdateLicenseSession(c)))
	licensingHandler(api, "/license-sessions/{CLIENT_SESSION_ID:[A-Za-z0-9_-]{43}=}", http.MethodDelete, withAPI(licDeleteLicenseSession(c)))

	// Resource API
	resourceHandler(api, "/license-issuers", http.MethodPost, withAPIAuthorized(createLicenseIssuer(c)))
	resourceHandler(api, "/license-issuers", http.MethodGet, withAPIAuthorized(getAllLicenseIssuers(c)))
	resourceHandler(api, "/license-issuers/{LICENSE_ISSUER_ID:[0-9]+}", http.MethodGet, withAPIAuthorized(getLicenseIssuer(c)))
	resourceHandler(api, "/license-issuers/{LICENSE_ISSUER_ID:[0-9]+}", http.MethodPatch, withAPIAuthorized(updateLicenseIssuer(c)))
	resourceHandler(api, "/license-issuers/{LICENSE_ISSUER_ID:[0-9]+}", http.MethodDelete, withAPIAuthorized(deleteLicenseIssuer(c)))

	apili := api.PathPrefix("/license-issuers/{LICENSE_ISSUER_ID:[0-9]+}").Subrouter()
	resourceHandler(apili, "/licenses", http.MethodPost, withAPIAuthorized(createLicense(c)))
	resourceHandler(apili, "/licenses", http.MethodGet, withAPIAuthorized(getAllLicenses(c)))
	resourceHandler(apili, "/licenses/{LICENSE_ID:[A-Za-z0-9_-]{43}=}", http.MethodGet, withAPIAuthorized(getLicense(c)))
	resourceHandler(apili, "/licenses/{LICENSE_ID:[A-Za-z0-9_-]{43}=}", http.MethodPatch, withAPIAuthorized(updateLicense(c)))
	resourceHandler(apili, "/licenses/{LICENSE_ID:[A-Za-z0-9_-]{43}=}", http.MethodDelete, withAPIAuthorized(deleteLicense(c)))

	resourceHandler(apili, "/products", http.MethodPost, withAPIAuthorized(createProduct(c)))
	resourceHandler(apili, "/products", http.MethodGet, withAPIAuthorized(getAllProducts(c)))
	resourceHandler(apili, "/products/{PRODUCT_ID:[0-9]+}", http.MethodGet, withAPIAuthorized(getProduct(c)))
	resourceHandler(apili, "/products/{PRODUCT_ID:[0-9]+}", http.MethodPatch, withAPIAuthorized(updateProduct(c)))
	resourceHandler(apili, "/products/{PRODUCT_ID:[0-9]+}", http.MethodDelete, withAPIAuthorized(deleteProduct(c)))

	apilil := apili.PathPrefix("/licenses/{LICENSE_ID:[A-Za-z0-9_-]{43}=}").Subrouter()
	resourceHandler(apilil, "/sessions", http.MethodGet, withAPIAuthorized(getAllLicenseSessions(c)))
	resourceHandler(apilil, "/sessions/{CLIENT_SESSION_ID:[A-Za-z0-9_-]{43}=}", http.MethodGet, withAPIAuthorized(getLicenseSession(c)))
	resourceHandler(apilil, "/sessions/{CLIENT_SESSION_ID:[A-Za-z0-9_-]{43}=}", http.MethodDelete, withAPIAuthorized(deleteLicenseSession(c)))

	// Auth API
	resourceHandler(api, "/login", http.MethodPost, withAPI(createToken(c)))
	resourceHandler(api, "/change-password/{LICENSE_ISSUER_ID:[0-9]+}", http.MethodPatch, withAPIAuthorized(updatePassword(c)))

	// Single page app
	if c.UseGUI() {
		r.PathPrefix("/").Handler(spaHandler(publicDir, "public"))
	}

	return r
}

func pathVarKey(str string) ([]byte, error) {
	if str == "" {
		return nil, errors.New("missing var")
	}
	bs, err := base64.URLEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}
	if len(bs) != 32 {
		return nil, errors.New("invalid length var")
	}
	return bs, nil
}
