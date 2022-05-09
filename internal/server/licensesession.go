package server

import (
	"bytes"
	"errors"
	"net/http"
	"strconv"
	"time"

	cryptorand "crypto/rand"

	"github.com/apex/log"
	"github.com/gorilla/mux"
	"github.com/sewiti/licensing-system/internal/core"
	"github.com/sewiti/licensing-system/internal/model"
	"github.com/sewiti/licensing-system/pkg/util"
)

// Licensing

func licCreateLicenseSession(c *core.Core) apiHandler {
	type createLicenseSessionReq struct {
		LicenseID []byte `json:"lid"`
		Data      []byte `json:"data"`
		N         []byte `json:"n"`
	}
	type createLicenseSessionReqData struct {
		ClientSessionID []byte    `json:"csid"`
		Identifier      string    `json:"id"`
		MachineID       []byte    `json:"machineID"`
		AppVersion      string    `json:"appVersion"`
		Timestamp       time.Time `json:"ts"`
	}
	type createLicenseSessionRes struct {
		Data []byte `json:"data"`
		N    []byte `json:"n"`
	}
	type createLicenseSessionResData struct {
		ServerSessionID []byte    `json:"ssid"`
		Timestamp       time.Time `json:"ts"`
		RefreshAfter    time.Time `json:"refresh"`
		ExpireAfter     time.Time `json:"expire"`
		Name            string    `json:"name,omitempty"`
		Data            []byte    `json:"data,omitempty"`
		ProductID       *int      `json:"productID,omitempty"`
		ProductName     string    `json:"productName"`
		ProductData     []byte    `json:"productData,omitempty"`
	}

	return func(r *http.Request) *apiResponse {
		const scope = "create license session"

		var req createLicenseSessionReq
		err := jsonDecodeLim(r.Body, &req)
		if err != nil {
			return responseBadRequest(err)
		}

		l, err := c.GetLicense(r.Context(), req.LicenseID)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound()
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}

		var data createLicenseSessionReqData
		err = util.OpenJsonBox(&data, req.Data, req.N, l.ID, c.ServerKey())
		if err != nil {
			return responseBadRequest(err)
		}

		ls, p, refresh, err := c.NewLicenseSession(
			r.Context(),
			l,
			data.ClientSessionID,
			data.Identifier,
			data.MachineID,
			data.AppVersion,
			data.Timestamp,
		)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrTimeOutOfSync):
				return responseForbidden(err)
			case errors.Is(err, core.ErrLicenseExpired):
				return responseForbidden(err)
			case errors.Is(err, core.ErrLicenseInactive):
				return responseForbidden(err)
			case errors.Is(err, core.ErrProductInactive):
				return responseForbidden(err)
			case errors.Is(err, core.ErrLicenseIssuerDisabled):
				return responseForbidden(err)
			case errors.Is(err, core.ErrRateLimitReached):
				return responseConflict(err)
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}

		resData := createLicenseSessionResData{
			ServerSessionID: ls.ServerID,
			RefreshAfter:    refresh,
			ExpireAfter:     ls.Expire,
			Timestamp:       time.Now(),
			Name:            l.Name,
			Data:            l.Data,
			ProductID:       l.ProductID,
			ProductName:     p.Name,
			ProductData:     p.Data,
		}
		nonce, err := util.GenerateNonce(cryptorand.Reader)
		if err != nil {
			log.WithError(err).Error("generating nonce")
			return responseInternalServerError()
		}
		box, err := util.SealJsonBox(resData, nonce, ls.ClientID, c.ServerKey())
		if err != nil {
			log.WithError(err).Error("sealing json box")
			return responseInternalServerError()
		}
		return responseJson(http.StatusCreated, createLicenseSessionRes{
			Data: box,
			N:    nonce,
		})
	}
}

func licUpdateLicenseSession(c *core.Core) apiHandler {
	type updateLicenseSessionReq struct {
		Data []byte `json:"data"`
		N    []byte `json:"n"`
	}
	type updateLicenseSessionReqData struct {
		Timestamp time.Time `json:"ts"`
	}
	type updateLicenseSessionRes struct {
		Data []byte `json:"data"`
		N    []byte `json:"n"`
	}
	type updateLicenseSessionResData struct {
		Timestamp    time.Time `json:"ts"`
		RefreshAfter time.Time `json:"refresh"`
		ExpireAfter  time.Time `json:"expire"`
		Name         string    `json:"name,omitempty"`
		Data         []byte    `json:"data,omitempty"`
		ProductID    *int      `json:"productID,omitempty"`
		ProductName  string    `json:"productName"`
		ProductData  []byte    `json:"productData,omitempty"`
	}

	return func(r *http.Request) *apiResponse {
		const scope = "update license session"
		clientSessionID, err := pathVarKey(mux.Vars(r)["CLIENT_SESSION_ID"])
		if err != nil {
			return responseBadRequestf("client session id: %v", err)
		}

		var req updateLicenseSessionReq
		err = jsonDecodeLim(r.Body, &req)
		if err != nil {
			return responseBadRequest(err)
		}
		ls, err := c.GetLicenseSession(r.Context(), clientSessionID)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound()
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}
		var reqData updateLicenseSessionReqData
		err = util.OpenJsonBox(&reqData, req.Data, req.N, ls.ClientID, ls.ServerKey)
		if err != nil {
			return responseBadRequest(err)
		}

		l, err := c.GetLicense(r.Context(), ls.LicenseID)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound()
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}
		p, refresh, err := c.UpdateLicenseSession(r.Context(), ls, l, reqData.Timestamp)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrTimeOutOfSync):
				return responseForbidden(err)
			case errors.Is(err, core.ErrLicenseExpired):
				return responseForbidden(err)
			case errors.Is(err, core.ErrLicenseInactive):
				return responseForbidden(err)
			case errors.Is(err, core.ErrProductInactive):
				return responseForbidden(err)
			case errors.Is(err, core.ErrLicenseIssuerDisabled):
				return responseForbidden(err)
			case errors.Is(err, core.ErrLicenseSessionExpired):
				return responseForbidden(err)
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound()
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}

		resData := updateLicenseSessionResData{
			Timestamp:    time.Now(),
			RefreshAfter: refresh,
			ExpireAfter:  ls.Expire,
			Name:         l.Name,
			Data:         l.Data,
			ProductID:    l.ProductID,
			ProductName:  p.Name,
			ProductData:  p.Data,
		}
		nonce, err := util.GenerateNonce(cryptorand.Reader)
		if err != nil {
			log.WithError(err).Error("generating nonce")
			return responseInternalServerError()
		}
		box, err := util.SealJsonBox(resData, nonce, ls.ClientID, ls.ServerKey)
		if err != nil {
			log.WithError(err).Error("sealing json box")
			return responseInternalServerError()
		}
		return responseJson(http.StatusCreated, updateLicenseSessionRes{
			Data: box,
			N:    nonce,
		})
	}
}

func licDeleteLicenseSession(c *core.Core) apiHandler {
	type deleteLicenseSessionReq struct {
		Data []byte `json:"data"`
		N    []byte `json:"n"`
	}
	type deleteLicenseSessionReqData struct {
		Timestamp time.Time `json:"ts"`
	}

	return func(r *http.Request) *apiResponse {
		const scope = "delete license session"
		clientSessionID, err := pathVarKey(mux.Vars(r)["CLIENT_SESSION_ID"])
		if err != nil {
			return responseBadRequestf("client session id: %v", err)
		}

		var req deleteLicenseSessionReq
		err = jsonDecodeLim(r.Body, &req)
		if err != nil {
			return responseBadRequest(err)
		}
		ls, err := c.GetLicenseSession(r.Context(), clientSessionID)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound()
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}
		var reqData deleteLicenseSessionReqData
		err = util.OpenJsonBox(&reqData, req.Data, req.N, ls.ClientID, ls.ServerKey)
		if err != nil {
			return responseBadRequest(err)
		}

		err = c.DeleteLicenseSession(r.Context(), ls.ClientID)
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

// Resource

func getAllLicenseSessions(c *core.Core) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "get all license sessions"
		licenseID, err := pathVarKey(mux.Vars(r)["LICENSE_ID"])
		if err != nil {
			return responseBadRequestf("license id: %v", err)
		}

		_, err = c.GetLicense(r.Context(), licenseID)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound()
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}

		lss, err := c.GetAllLicenseSessionsByLicense(r.Context(), licenseID)
		if err != nil {
			logError(err, scope)
			return responseInternalServerError()
		}
		if lss == nil {
			lss = make([]*model.LicenseSession, 0) // Force empty array json
		}
		return responseJson(http.StatusOK, lss)
	}
}

func getLicenseSession(c *core.Core) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "get license session"
		vars := mux.Vars(r)
		licenseID, err := pathVarKey(vars["LICENSE_ID"])
		if err != nil {
			return responseBadRequestf("license id: %v", err)
		}
		clientSessionID, err := pathVarKey(vars["CLIENT_SESSION_ID"])
		if err != nil {
			return responseBadRequestf("client session id: %v", err)
		}
		// TODO: License is useless

		ls, err := c.GetLicenseSession(r.Context(), clientSessionID)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound()
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}
		if !bytes.Equal(licenseID, ls.LicenseID) {
			return responseNotFound()
		}
		return responseJson(http.StatusOK, ls)
	}
}

func deleteLicenseSession(c *core.Core) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "delete license session"
		vars := mux.Vars(r)
		licenseIssuerID, err := strconv.Atoi(vars["LICENSE_ISSUER_ID"])
		if err != nil {
			return responseBadRequestf("license issuer id: %v", err)
		}
		licenseID, err := pathVarKey(vars["LICENSE_ID"])
		if err != nil {
			return responseBadRequestf("license id: %v", err)
		}
		clientSessionID, err := pathVarKey(vars["CLIENT_SESSION_ID"])
		if err != nil {
			return responseBadRequestf("client session id: %v", err)
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
		if l.IssuerID != licenseIssuerID {
			return responseNotFound()
		}

		err = c.DeleteLicenseSession(r.Context(), clientSessionID)
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
