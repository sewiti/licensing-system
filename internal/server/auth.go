package server

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sewiti/licensing-system/internal/core"
	"github.com/sewiti/licensing-system/internal/core/auth"
	"github.com/sewiti/licensing-system/internal/model"
)

type apiAuthHandler func(r *http.Request, login *model.LicenseIssuer) *apiResponse

func withAPIAuth(c *core.Core, h apiAuthHandler) apiHandler {
	bearerAuth := func(authHeader string) (token string, ok bool) {
		const prefix = "bearer "
		if !strings.HasPrefix(strings.ToLower(authHeader), prefix) {
			return "", false
		}
		return authHeader[len(prefix):], true
	}

	return func(r *http.Request) *apiResponse {
		// Basic auth
		username, passwd, ok := r.BasicAuth()
		if ok {
			li, err := c.AuthenticateBasic(r.Context(), username, passwd)
			if err != nil {
				switch {
				case errors.Is(err, core.ErrNotFound):
					return responseUnauthorized()
				case errors.Is(err, core.ErrUserInactive):
					return responseUnauthorized()
				case errors.Is(err, auth.ErrInvalidPasswd):
					return responseUnauthorized()
				case errors.Is(err, auth.ErrNoLogin):
					return responseUnauthorized()
				default:
					logError(err, "basic-auth")
					return responseInternalServerError()
				}
			}
			return h(r, li)
		}

		// Bearer token
		token, ok := bearerAuth(r.Header.Get("Authorization"))
		if ok {
			li, err := c.AuthenticateToken(r.Context(), token)
			if err != nil {
				switch {
				case errors.Is(err, core.ErrNotFound):
					return responseUnauthorized()
				case errors.Is(err, core.ErrUserInactive):
					return responseUnauthorized()
				default:
					logError(err, "bearer-auth")
					return responseInternalServerError()
				}
			}
			return h(r, li)
		}
		return responseUnauthorized()
	}
}

func withAuthorized(c *core.Core, h apiAuthHandler) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		licenseIssuerIDStr, ok := mux.Vars(r)["LICENSE_ISSUER_ID"]
		if ok {
			licenseIssuerID, err := strconv.Atoi(licenseIssuerIDStr)
			if err != nil {
				return responseBadRequest(err)
			}
			if licenseIssuerID == login.ID {
				// ok - authorized for self
				return h(r, login)
			}
		}

		if c.IsPrivileged(login) {
			// ok - privileged user
			return h(r, login)
		}
		return responseForbidden()
	}
}

func createToken(c *core.Core) apiHandler {
	type createTokenReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	type createTokenRes struct {
		Token string `json:"token"`
	}

	return func(r *http.Request) *apiResponse {
		const scope = "create token"
		var req createTokenReq
		err := jsonDecodeLim(r.Body, &req)
		if err != nil {
			return responseBadRequest(err)
		}

		li, err := c.AuthenticateBasic(r.Context(), req.Username, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrNotFound):
				return responseUnauthorized()
			case errors.Is(err, core.ErrUserInactive):
				return responseUnauthorized()
			case errors.Is(err, auth.ErrInvalidPasswd):
				return responseUnauthorized()
			case errors.Is(err, auth.ErrNoLogin):
				return responseUnauthorized()
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}
		token, err := c.CreateToken(li)
		if err != nil {
			logError(err, scope)
			return responseInternalServerError()
		}
		return responseJson(http.StatusOK, createTokenRes{
			Token: token,
		})
	}
}

func updatePassword(c *core.Core) apiAuthHandler {
	type updatePasswordReq struct {
		OldPasswd string `json:"oldPassword"`
		NewPasswd string `json:"newPassword"`
	}

	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "update password"
		licenseIssuerID, err := strconv.Atoi(mux.Vars(r)["LICENSE_ISSUER_ID"])
		if err != nil {
			return responseBadRequestf("license issuer id: %v", err)
		}
		var req updatePasswordReq
		err = jsonDecodeLim(r.Body, &req)
		if err != nil {
			return responseBadRequest(err)
		}

		err = c.ChangePasswd(r.Context(), login, licenseIssuerID, req.OldPasswd, req.NewPasswd)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound()
			case errors.Is(err, auth.ErrInvalidPasswd):
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
