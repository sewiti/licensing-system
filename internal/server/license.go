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

func createLicense(c *core.Core) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "create license"
		licenseIssuerID, err := strconv.Atoi(mux.Vars(r)["LICENSE_ISSUER_ID"])
		if err != nil {
			return responseBadRequestf("license issuer id: %v", err)
		}

		req := model.License{ // only a handful of fields will be used
			Active:      true,
			MaxSessions: 1,
		}
		err = jsonDecodeLim(r.Body, &req)
		if err != nil {
			return responseBadRequest(err)
		}
		if req.Tags == nil {
			req.Tags = make([]string, 0)
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

		l, err := c.NewLicense(r.Context(), li, &req)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrInvalidInput):
				return responseBadRequest(err)
			case errors.Is(err, core.ErrExceedsLimit):
				return responseBadRequest(err)
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}
		return responseJson(http.StatusCreated, l)
	}
}

func getAllLicenses(c *core.Core) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "get all licenses"
		licenseIssuerID, err := strconv.Atoi(mux.Vars(r)["LICENSE_ISSUER_ID"])
		if err != nil {
			return responseBadRequestf("license issuer id: %v", err)
		}

		_, err = c.GetLicenseIssuer(r.Context(), licenseIssuerID)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound()
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}

		ll, err := c.GetAllLicensesByIssuer(r.Context(), licenseIssuerID)
		if err != nil {
			logError(err, scope)
			return responseInternalServerError()
		}
		if ll == nil {
			ll = make([]*model.License, 0) // Force empty array json
		}
		return responseJson(http.StatusOK, ll)
	}
}

func getLicense(c *core.Core) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "get license"
		vars := mux.Vars(r)
		licenseIssuerID, err := strconv.Atoi(vars["LICENSE_ISSUER_ID"])
		if err != nil {
			return responseBadRequestf("license issuer id: %v", err)
		}
		licenseID, err := pathVarKey(vars["LICENSE_ID"])
		if err != nil {
			return responseBadRequestf("license id: %v", err)
		}

		l, err := c.GetLicense(r.Context(), licenseID)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound()
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}
		if licenseIssuerID != l.IssuerID {
			return responseNotFound()
		}
		return responseJson(http.StatusOK, l)
	}
}

func updateLicense(c *core.Core) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "update license"
		vars := mux.Vars(r)
		licenseIssuerID, err := strconv.Atoi(vars["LICENSE_ISSUER_ID"])
		if err != nil {
			return responseBadRequestf("license issuer id: %v", err)
		}
		licenseID, err := pathVarKey(vars["LICENSE_ID"])
		if err != nil {
			return responseBadRequestf("license id: %v", err)
		}

		data, err := readAllLim(r.Body)
		if err != nil {
			return responseBadRequest(err)
		}
		l := &model.License{
			ID:       licenseID,
			IssuerID: licenseIssuerID,
		}
		err = json.Unmarshal(data, l)
		if err != nil {
			return responseBadRequest(err)
		}

		changes, err := core.UnmarshalChanges(data)
		if err != nil {
			return responseBadRequest(err) // should never happen
		}
		mask, _ := c.AuthorizeLicenseUpdate(login)
		field, ok := core.ChangesInMask(changes, mask)
		if !ok {
			return responseBadRequestf("unauthorized to change field: %s", field)
		}

		err = c.UpdateLicense(r.Context(), l, changes)
		if err != nil {
			// TODO
			switch {
			case errors.Is(err, core.ErrInvalidInput):
				return responseBadRequest(err)
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}

		l, err = c.GetLicense(r.Context(), licenseID)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound() // should never happen
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}
		return responseJson(http.StatusOK, l)
	}
}

func deleteLicense(c *core.Core) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "delete license"
		vars := mux.Vars(r)
		licenseIssuerID, err := strconv.Atoi(vars["LICENSE_ISSUER_ID"])
		if err != nil {
			return responseBadRequestf("license issuer id: %v", err)
		}
		licenseID, err := pathVarKey(vars["LICENSE_ID"])
		if err != nil {
			return responseBadRequestf("license id: %v", err)
		}

		_, canDelete := c.AuthorizeLicenseUpdate(login)
		if !canDelete {
			return responseForbidden()
		}
		err = c.DeleteLicense(r.Context(), licenseID, licenseIssuerID)
		if err != nil {
			switch {
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
