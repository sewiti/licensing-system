package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sewiti/licensing-system/internal/core"
	"github.com/sewiti/licensing-system/internal/model"
)

func createLicenseIssuer(c *core.Core) apiAuthHandler {
	type createLicenseIssuerReq struct {
		Password            string `json:"password"`
		model.LicenseIssuer        // only a handful of fields are used
	}

	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "create license issuer"
		var req createLicenseIssuerReq
		err := jsonDecodeLim(r.Body, &req)
		if err != nil {
			return responseBadRequest(err)
		}

		li, err := c.NewLicenseIssuer(r.Context(), req.Username, req.Password, req.Email, req.PhoneNumber, req.MaxLicenses)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrInvalidInput):
				return responseBadRequest(err)
			case errors.Is(err, core.ErrPasswdTooWeak):
				return responseBadRequest(err)
			case errors.Is(err, core.ErrDuplicate):
				return responseConflict(err)
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}
		return responseJson(http.StatusCreated, li)
	}
}

func getAllLicenseIssuers(c *core.Core) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "get all license issuers"
		lii, err := c.GetAllLicenseIssuers(r.Context())
		if err != nil {
			logError(err, scope)
			return responseInternalServerError()
		}
		if lii == nil {
			lii = make([]*model.LicenseIssuer, 0) // Force empty array json
		}
		return responseJson(http.StatusOK, lii)
	}
}

func getLicenseIssuer(c *core.Core) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "get license issuer"
		licenseIssuerID, err := strconv.Atoi(mux.Vars(r)["LICENSE_ISSUER_ID"])
		if err != nil {
			return responseBadRequestf("license issuer id: %v", err)
		}

		li, err := c.GetLicenseIssuer(r.Context(), licenseIssuerID)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound()
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}
		return responseJson(http.StatusOK, li)
	}
}

func updateLicenseIssuer(c *core.Core) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "update license issuer"
		licenseIssuerID, err := strconv.Atoi(mux.Vars(r)["LICENSE_ISSUER_ID"])
		if err != nil {
			return responseBadRequestf("license issuer id: %v", err)
		}

		data, err := readAllLim(r.Body)
		if err != nil {
			return responseBadRequest(err)
		}
		li := &model.LicenseIssuer{
			ID: licenseIssuerID,
		}
		err = json.Unmarshal(data, li)
		if err != nil {
			return responseBadRequest(err)
		}

		changes, err := core.UnmarshalChanges(data)
		if err != nil {
			return responseBadRequest(err) // should never happen
		}
		mask, _ := c.AuthorizeLicenseIssuerUpdate(login)
		field, ok := core.ChangesInMask(changes, mask)
		if !ok {
			return responseBadRequestf("unauthorized to change field: %s", field)
		}

		err = c.UpdateLicenseIssuer(r.Context(), li, changes)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrSuperadminImmutable):
				return responseForbidden(err)
			case errors.Is(err, core.ErrDuplicate):
				return responseConflict(err)
			case errors.Is(err, core.ErrInvalidInput):
				return responseBadRequest(err)
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound()
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}

		li, err = c.GetLicenseIssuer(r.Context(), licenseIssuerID)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound() // should never happen
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}
		return responseJson(http.StatusOK, li)
	}
}

func deleteLicenseIssuer(c *core.Core) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "delete license issuer"
		licenseIssuerID, err := strconv.Atoi(mux.Vars(r)["LICENSE_ISSUER_ID"])
		if err != nil {
			return responseBadRequestf("license issuer id: %v", err)
		}

		_, canDelete := c.AuthorizeLicenseIssuerUpdate(login)
		if !canDelete {
			return responseForbidden()
		}
		err = c.DeleteLicenseIssuer(r.Context(), licenseIssuerID)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrSuperadminImmutable):
				return responseForbidden(err)
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
