package server

import (
	cryptorand "crypto/rand"
	"encoding/base64"
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
	MachineUUID     []byte    `json:"machineUUID"`
	Timestamp       time.Time `json:"ts"`
}

type createLicenseSessionRes struct {
	Data []byte    `json:"data"`
	N    *[24]byte `json:"n"`
}

type createLicenseSessionResData struct {
	ServerSessionID *[32]byte   `json:"ssid"`
	Timestamp       time.Time   `json:"ts"`
	RefreshAfter    time.Time   `json:"refresh"`
	StopAfter       time.Time   `json:"stop"`
	Data            interface{} `json:"data,omitempty"`
}

func createLicenseSession(c *core.Core) apiHandler {
	const scope = "create license-session"
	return func(r *http.Request) *jsonResponse {
		var req createLicenseSessionReq
		err := jsonDecodeLim(r.Body, &req)
		if err != nil {
			return respondBadRequest(err)
		}
		l, err := c.GetLicense(r.Context(), req.LicenseID)
		if err != nil {
			return errNotFound
		}

		var data createLicenseSessionReqData
		err = util.OpenJsonBox(&data, req.Data, req.N, l.ID, c.ServerKey())
		if err != nil {
			return respondBadRequest(err)
		}

		ls, refresh, err := c.NewLicenseSession(r.Context(), l, data.ClientSessionID, data.MachineUUID, data.Timestamp)
		if err != nil {
			if err, ok := err.(*core.SensitiveError); ok {
				log.WithError(err.Unwrap()).Errorf("%s: %s", scope, err.Message)
				return errInternalServerError
			}
			return respondBadRequest(err.Error())
		}

		resData := createLicenseSessionResData{
			ServerSessionID: ls.ServerID,
			RefreshAfter:    refresh,
			StopAfter:       ls.Expire,
			Timestamp:       time.Now(),
		}
		nonce, err := util.GenerateNonce(cryptorand.Reader)
		if err != nil {
			log.WithError(err).Error("generating nonce")
			return errInternalServerError
		}
		box, err := util.SealJsonBox(resData, nonce, ls.ClientID, c.ServerKey())
		if err != nil {
			log.WithError(err).Error("sealing json box")
			return errInternalServerError
		}
		return respondJson(http.StatusCreated, createLicenseSessionRes{
			Data: box,
			N:    nonce,
		})
	}
}

type updateLicenseSessionReq struct {
	ClientSessionID *[32]byte `json:"csid"`
	Data            []byte    `json:"data"`
	N               *[24]byte `json:"n"`
}

type updateLicenseSessionReqData struct {
	Timestamp time.Time `json:"ts"`
}

type updateLicenseSessionRes struct {
	Data []byte    `json:"data"`
	N    *[24]byte `json:"n"`
}

type updateLicenseSessionResData struct {
	Timestamp    time.Time   `json:"ts"`
	RefreshAfter time.Time   `json:"refresh"`
	StopAfter    time.Time   `json:"stop"`
	Data         interface{} `json:"data,omitempty"`
}

func updateLicenseSession(c *core.Core) apiHandler {
	const scope = "update license-session"
	return func(r *http.Request) *jsonResponse {
		clientSessionIDStr, ok := mux.Vars(r)["CSID"]
		if !ok {
			return respondBadRequest("missing client session id")
		}
		clientSessionID, err := base64.URLEncoding.DecodeString(clientSessionIDStr)
		if err != nil {
			return respondBadRequestf("client session id: %v", err)
		}

		var req updateLicenseSessionReq
		err = jsonDecodeLim(r.Body, &req)
		if err != nil {
			return respondBadRequest(err)
		}
		ls, err := c.GetLicenseSession(r.Context(), (*[32]byte)(clientSessionID))
		if err != nil {
			return errNotFound
		}
		var data updateLicenseSessionReqData
		err = util.OpenJsonBox(&data, req.Data, req.N, ls.ClientID, ls.ServerKey)
		if err != nil {
			return respondBadRequest(err)
		}

		l, err := c.GetLicense(r.Context(), ls.LicenseID)
		if err != nil {
			log.WithError(err).Errorf("%s: getting license", scope)
			return errInternalServerError
		}
		refresh, err := c.UpdateLicenseSession(r.Context(), ls, l, data.Timestamp)
		if err != nil {
			if _, ok := err.(*core.SensitiveError); ok {
				log.WithError(err).Errorf("%s: updating license session", scope)
				return errInternalServerError
			}
			return errInternalServerError
		}

		resData := updateLicenseSessionResData{
			Timestamp:    time.Now(),
			RefreshAfter: refresh,
			StopAfter:    ls.Expire,
			Data:         l.Data,
		}
		nonce, err := util.GenerateNonce(cryptorand.Reader)
		if err != nil {
			log.WithError(err).Error("generating nonce")
			return errInternalServerError
		}
		box, err := util.SealJsonBox(resData, nonce, ls.ClientID, ls.ServerKey)
		if err != nil {
			log.WithError(err).Error("sealing json box")
			return errInternalServerError
		}
		return respondJson(http.StatusCreated, updateLicenseSessionRes{
			Data: box,
			N:    nonce,
		})
	}
}

type deleteLicenseSessionReq struct {
	ClientSessionID *[32]byte `json:"csid"`
	Data            []byte    `json:"data"`
	N               *[24]byte `json:"n"`
}

type deleteLicenseSessionReqData struct {
	Uptime    time.Duration `json:"uptime"`
	Timestamp time.Time     `json:"ts"`
}

func deleteLicenseSession(c *core.Core) apiHandler {
	return func(r *http.Request) *jsonResponse {
		clientSessionIDStr, ok := mux.Vars(r)["CSID"] // client session id
		if !ok {
			return respondBadRequest("missing client session id")
		}
		clientSessionID, err := base64.URLEncoding.DecodeString(clientSessionIDStr)
		if err != nil {
			return respondBadRequestf("client session id: %v", err)
		}

		var req deleteLicenseSessionReq
		err = jsonDecodeLim(r.Body, &req)
		if err != nil {
			return respondBadRequest(err)
		}
		ls, err := c.GetLicenseSession(r.Context(), (*[32]byte)(clientSessionID))
		if err != nil {
			return errNotFound
		}
		var data deleteLicenseSessionReqData
		err = util.OpenJsonBox(&data, req.Data, req.N, ls.ClientID, ls.ServerKey)
		if err != nil {
			return respondBadRequest(err)
		}

		err = c.DeleteLicenseSession(r.Context(), ls.ClientID)
		if err != nil {
			return errNotFound
		}
		return respondNoContent
	}
}
