package server

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
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
			MaxSessions: 1,
		}
		err = jsonDecodeLim(r.Body, &req)
		if err != nil {
			return responseBadRequest(err)
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

		l, err := c.NewLicense(r.Context(), li, req.Note, req.Data, req.MaxSessions, req.ValidUntil)
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

		ll, err := c.GetAllLicensesByIssuer(r.Context(), licenseIssuerID)
		if err != nil {
			logError(err, scope)
			return responseInternalServerError()
		}
		return responseJson(http.StatusOK, ll)
	}
}

func getLicense(c *core.Core) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "get license"
		vars := mux.Vars(r)
		_, err := strconv.Atoi(vars["LICENSE_ISSUER_ID"])
		if err != nil {
			return responseBadRequestf("license issuer id: %v", err)
		}
		licenseID, err := pathVarID(vars["LICENSE_ID"])
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
		return responseJson(http.StatusOK, l)
	}
}

func updateLicense(c *core.Core) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "update license"
		vars := mux.Vars(r)
		_, err := strconv.Atoi(vars["LICENSE_ISSUER_ID"])
		if err != nil {
			return responseBadRequestf("license issuer id: %v", err)
		}
		licenseID, err := pathVarID(vars["LICENSE_ID"])
		if err != nil {
			return responseBadRequestf("license id: %v", err)
		}

		data, err := ioutil.ReadAll(io.LimitReader(r.Body, maxRequestSize))
		if err != nil {
			return responseBadRequest(err)
		}
		{
			// Validate schema
			var li model.License
			err = json.Unmarshal(data, &li)
			if err != nil {
				return responseBadRequest(err)
			}
		}

		update := make(map[string]interface{})
		err = json.Unmarshal(data, &update)
		if err != nil {
			return responseBadRequest(err) // should never happen
		}
		mask, _ := c.AuthorizeLicenseUpdate(login)
		field, ok := core.UpdateInMask(update, mask)
		if !ok {
			return responseBadRequestf("unable to change field: %s", field)
		}

		err = c.UpdateLicense(r.Context(), licenseID, update)
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

		l, err := c.GetLicense(r.Context(), licenseID)
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
		_, err := strconv.Atoi(vars["LICENSE_ISSUER_ID"])
		if err != nil {
			return responseBadRequestf("license issuer id: %v", err)
		}
		licenseID, err := pathVarID(vars["LICENSE_ID"])
		if err != nil {
			return responseBadRequestf("license id: %v", err)
		}

		_, canDelete := c.AuthorizeLicenseUpdate(login)
		if !canDelete {
			return responseForbidden()
		}
		err = c.DeleteLicense(r.Context(), licenseID)
		if err != nil {
			logError(err, scope)
			return responseInternalServerError()
		}
		return responseNoContent()
	}
}
