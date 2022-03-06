package server

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sewiti/licensing-system/internal/core"
	"github.com/sewiti/licensing-system/internal/core/auth"
)

func NewRouterInternal(c *core.Core) *mux.Router {
	r := mux.NewRouter()

	r.Path("/license-issuers/{LICENSE_ISSUER_USERNAME:[A-Za-z0-9_-]+}/active").
		Methods(http.MethodPatch).Handler(withAPI(internalUpdateLicenseIssuerActive(c)))

	r.Path("/license-issuers/{LICENSE_ISSUER_USERNAME:[A-Za-z0-9_-]+}/change-password").
		Methods(http.MethodPatch).Handler(withAPI(internalUpdateLicenseIssuerPassword(c)))

	return r
}

func internalUpdateLicenseIssuerActive(c *core.Core) apiHandler {
	type updateLicenseIssuerActiveReq struct {
		Active bool `json:"active"`
	}

	return func(r *http.Request) *apiResponse {
		const scope = "internal update license issuer active"
		username, ok := mux.Vars(r)["LICENSE_ISSUER_USERNAME"]
		if !ok {
			return responseBadRequestf("license issuer username: missing")
		}
		var req updateLicenseIssuerActiveReq
		err := jsonDecodeLim(r.Body, &req)
		if err != nil {
			return responseBadRequest(err)
		}

		li, err := c.GetLicenseIssuerByUsername(r.Context(), username)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound()
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}

		err = c.UpdateLicenseIssuerBypass(r.Context(), li.ID, map[string]interface{}{
			"active": req.Active,
		})
		if err != nil {
			switch {
			case errors.Is(err, core.ErrInvalidInput):
				return responseBadRequest(err)
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound()
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}
		return responseNoContent()
	}
}

func internalUpdateLicenseIssuerPassword(c *core.Core) apiHandler {
	type updateLicenseIssuerPassword struct {
		NewPassword string `json:"newPassword"`
	}

	return func(r *http.Request) *apiResponse {
		const scope = "internal update license issuer password"
		username, ok := mux.Vars(r)["LICENSE_ISSUER_USERNAME"]
		if !ok {
			return responseBadRequestf("license issuer username: missing")
		}
		var req updateLicenseIssuerPassword
		err := jsonDecodeLim(r.Body, &req)
		if err != nil {
			return responseBadRequest(err)
		}

		li, err := c.GetLicenseIssuerByUsername(r.Context(), username)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound()
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}

		err = c.ChangePasswd(r.Context(), core.CLILogin(), li.ID, "", req.NewPassword)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound()
			case errors.Is(err, auth.ErrInvalidPasswd): // should never happen
				return responseForbidden(err)
			case errors.Is(err, core.ErrPasswdTooWeak):
				return responseBadRequest(err)
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}
		return responseNoContent()
	}
}
