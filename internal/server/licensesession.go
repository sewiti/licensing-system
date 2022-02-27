package server

import (
	cryptorand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/gorilla/mux"
	"github.com/sewiti/licensing-system/internal/core"
	"github.com/sewiti/licensing-system/pkg/util"
)

type createLicenseSessionReq struct {
	LicenseID *[32]byte `json:"lid"`
	Data      []byte    `json:"data"`
	N         *[24]byte `json:"n"`
}

type createLicenseSessionReqData struct {
	ClientSessionID *[32]byte `json:"csid"`
	MachineID       []byte    `json:"machineID"`
	Timestamp       time.Time `json:"ts"`
}

type createLicenseSessionRes struct {
	Data []byte    `json:"data"`
	N    *[24]byte `json:"n"`
}

type createLicenseSessionResData struct {
	ServerSessionID *[32]byte       `json:"ssid"`
	Timestamp       time.Time       `json:"ts"`
	RefreshAfter    time.Time       `json:"refresh"`
	ExpireAfter     time.Time       `json:"expire"`
	Data            json.RawMessage `json:"data,omitempty"`
}

func createLicenseSession(c *core.Core) apiHandler {
	const scope = "create license-session"
	return func(r *http.Request) *apiResponse {
		var req createLicenseSessionReq
		err := jsonDecodeLim(r.Body, &req)
		if err != nil {
			return responseBadRequest(err)
		}
		l, err := c.GetLicense(r.Context(), req.LicenseID)
		if err != nil {
			if err, ok := err.(*core.SensitiveError); ok {
				log.WithError(err.Unwrap()).Errorf("%s: %s", scope, err.Message)
			}
			if errors.Is(err, core.ErrNotFound) {
				return responseNotFound()
			}
			return responseInternalServerError()
		}

		var data createLicenseSessionReqData
		err = util.OpenJsonBox(&data, req.Data, req.N, l.ID, c.ServerKey())
		if err != nil {
			return responseBadRequest(err)
		}

		ls, refresh, err := c.NewLicenseSession(r.Context(), l, data.ClientSessionID, data.MachineID, data.Timestamp)
		if err != nil {
			if err, ok := err.(*core.SensitiveError); ok {
				log.WithError(err.Unwrap()).Errorf("%s: %s", scope, err.Message)
			}
			switch {
			case errors.Is(err, core.ErrRateLimitReached):
				return responseConflict(err)
			case errors.Is(err, core.ErrLicenseExpired):
				return responseForbidden(err)
			case errors.Is(err, core.ErrTimeOutOfSync):
				return responseForbidden(err)
			default:
				return responseInternalServerError()
			}
		}

		resData := createLicenseSessionResData{
			ServerSessionID: ls.ServerID,
			RefreshAfter:    refresh,
			ExpireAfter:     ls.Expire,
			Timestamp:       time.Now(),
			Data:            l.Data,
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

type updateLicenseSessionReq struct {
	Data []byte    `json:"data"`
	N    *[24]byte `json:"n"`
}

type updateLicenseSessionReqData struct {
	Timestamp time.Time `json:"ts"`
}

type updateLicenseSessionRes struct {
	Data []byte    `json:"data"`
	N    *[24]byte `json:"n"`
}

type updateLicenseSessionResData struct {
	Timestamp    time.Time       `json:"ts"`
	RefreshAfter time.Time       `json:"refresh"`
	ExpireAfter  time.Time       `json:"expire"`
	Data         json.RawMessage `json:"data,omitempty"`
}

func updateLicenseSession(c *core.Core) apiHandler {
	const scope = "update license-session"
	return func(r *http.Request) *apiResponse {
		clientSessionIDStr, ok := mux.Vars(r)["CSID"]
		if !ok {
			return responseBadRequest("missing client session id")
		}
		clientSessionID, err := base64.URLEncoding.DecodeString(clientSessionIDStr)
		if err != nil {
			return responseBadRequestf("client session id: %v", err)
		}

		var req updateLicenseSessionReq
		err = jsonDecodeLim(r.Body, &req)
		if err != nil {
			return responseBadRequest(err)
		}
		ls, err := c.GetLicenseSession(r.Context(), (*[32]byte)(clientSessionID))
		if err != nil {
			if err, ok := err.(*core.SensitiveError); ok {
				log.WithError(err.Unwrap()).Errorf("%s: %s", scope, err.Message)
			}
			if errors.Is(err, core.ErrNotFound) {
				return responseNotFound()
			}
			return responseInternalServerError()
		}
		var reqData updateLicenseSessionReqData
		err = util.OpenJsonBox(&reqData, req.Data, req.N, ls.ClientID, ls.ServerKey)
		if err != nil {
			return responseBadRequest(err)
		}

		l, err := c.GetLicense(r.Context(), ls.LicenseID)
		if err != nil {
			if err, ok := err.(*core.SensitiveError); ok {
				log.WithError(err.Unwrap()).Errorf("%s: %s", scope, err.Message)
			}
			if errors.Is(err, core.ErrNotFound) {
				return responseNotFound()
			}
			return responseInternalServerError()
		}
		refresh, err := c.UpdateLicenseSession(r.Context(), ls, l, reqData.Timestamp)
		if err != nil {
			if err, ok := err.(*core.SensitiveError); ok {
				log.WithError(err).Errorf("%s: updating license session", scope)
			}
			switch {
			case errors.Is(err, core.ErrLicenseSessionExpired):
				return responseForbidden(err)
			case errors.Is(err, core.ErrLicenseExpired):
				return responseForbidden(err)
			case errors.Is(err, core.ErrTimeOutOfSync):
				return responseForbidden(err)
			default:
				return responseInternalServerError()
			}
		}

		resData := updateLicenseSessionResData{
			Timestamp:    time.Now(),
			RefreshAfter: refresh,
			ExpireAfter:  ls.Expire,
			Data:         l.Data,
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

type deleteLicenseSessionReq struct {
	Data []byte    `json:"data"`
	N    *[24]byte `json:"n"`
}

type deleteLicenseSessionReqData struct {
	Timestamp time.Time `json:"ts"`
}

func deleteLicenseSession(c *core.Core) apiHandler {
	const scope = "delete license-session"
	return func(r *http.Request) *apiResponse {
		clientSessionIDStr, ok := mux.Vars(r)["CSID"] // client session id
		if !ok {
			return responseBadRequest("missing client session id")
		}
		clientSessionID, err := base64.URLEncoding.DecodeString(clientSessionIDStr)
		if err != nil {
			return responseBadRequestf("client session id: %v", err)
		}

		var req deleteLicenseSessionReq
		err = jsonDecodeLim(r.Body, &req)
		if err != nil {
			return responseBadRequest(err)
		}
		ls, err := c.GetLicenseSession(r.Context(), (*[32]byte)(clientSessionID))
		if err != nil {
			if err, ok := err.(*core.SensitiveError); ok {
				log.WithError(err.Unwrap()).Errorf("%s: %s", scope, err.Message)
			}
			if errors.Is(err, core.ErrNotFound) {
				return responseNotFound()
			}
			return responseInternalServerError()
		}
		var reqData deleteLicenseSessionReqData
		err = util.OpenJsonBox(&reqData, req.Data, req.N, ls.ClientID, ls.ServerKey)
		if err != nil {
			return responseBadRequest(err)
		}

		err = c.DeleteLicenseSession(r.Context(), ls.ClientID)
		if err != nil {
			return responseNotFound()
		}
		return responseNoContent()
	}
}
