package license

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sewiti/licensing-system/pkg/util"
)

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

type deleteLicenseSessionReq struct {
	Data []byte `json:"data"`
	N    []byte `json:"n"`
}

type deleteLicenseSessionReqData struct {
	Timestamp time.Time `json:"ts"`
}

var errTemporary = errors.New("temporary")

func sendJsonRequest(ctx context.Context, method, url string, reqData, resData interface{}) error {
	bs, err := json.Marshal(reqData)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(bs))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", errTemporary, err)
	}
	defer r.Body.Close()

	switch r.StatusCode {
	case http.StatusOK, http.StatusCreated:
		if resData != nil {
			err = json.NewDecoder(r.Body).Decode(resData)
			if err != nil {
				return fmt.Errorf("%w: %v", errTemporary, err)
			}
		}
		return nil

	case http.StatusNoContent:
		return nil

	case http.StatusConflict:
		var msg struct {
			Message string `json:"message"`
		}
		err = json.NewDecoder(r.Body).Decode(&msg)
		if err != nil {
			return fmt.Errorf("%w: unexpected status: %s", errTemporary, r.Status)
		}
		return fmt.Errorf("%w: unexpected status: %s: %s", errTemporary, r.Status, msg.Message)

	case http.StatusNotFound:
		return fmt.Errorf("unexpected status: %s", r.Status)

	case http.StatusInternalServerError:
		return fmt.Errorf("%w: unexpected status: %s", errTemporary, r.Status)

	default:
		var msg struct {
			Message string `json:"message"`
		}
		err = json.NewDecoder(r.Body).Decode(&msg)
		if err != nil {
			return fmt.Errorf("unexpected status: %s", r.Status)
		}
		return fmt.Errorf("unexpected status: %s: %s", r.Status, msg.Message)
	}
}

func (c *Client) sendCreateSession(ctx context.Context, clientID, clientKey []byte, rand io.Reader) (*createLicenseSessionResData, error) {
	reqData := createLicenseSessionReqData{
		ClientSessionID: clientID,
		Identifier:      c.identifier,
		MachineID:       c.machineID,
		AppVersion:      c.appVersion,
		Timestamp:       time.Now(),
	}
	nonce, err := util.GenerateNonce(rand)
	if err != nil {
		return nil, err
	}
	bs, err := util.SealJsonBox(reqData, nonce, c.serverID, c.licenseKey)
	if err != nil {
		return nil, err
	}

	req := createLicenseSessionReq{
		LicenseID: c.licenseID,
		Data:      bs,
		N:         nonce,
	}
	var res createLicenseSessionRes
	err = sendJsonRequest(ctx, http.MethodPost, c.url, req, &res)
	if err != nil {
		return nil, err
	}

	var resData createLicenseSessionResData
	err = util.OpenJsonBox(&resData, res.Data, res.N, c.serverID, clientKey)
	if err != nil {
		return nil, err
	}
	return &resData, nil
}

func (s *session) sendRefresh(ctx context.Context, rand io.Reader) (*updateLicenseSessionResData, error) {
	reqData := updateLicenseSessionReqData{
		Timestamp: time.Now(),
	}
	nonce, err := util.GenerateNonce(rand)
	if err != nil {
		return nil, err
	}
	bs, err := util.SealJsonBox(reqData, nonce, s.serverID, s.clientKey)
	if err != nil {
		return nil, err
	}

	req := updateLicenseSessionReq{
		Data: bs,
		N:    nonce,
	}
	var res updateLicenseSessionRes
	url := fmt.Sprintf("%s/%s", s.url, base64.URLEncoding.EncodeToString(s.clientID))
	err = sendJsonRequest(ctx, http.MethodPatch, url, req, &res)
	if err != nil {
		return nil, err
	}

	var resData updateLicenseSessionResData
	err = util.OpenJsonBox(&resData, res.Data, res.N, s.serverID, s.clientKey)
	if err != nil {
		return nil, err
	}
	return &resData, nil
}

func (s *session) sendClose(ctx context.Context, rand io.Reader) error {
	reqData := deleteLicenseSessionReqData{
		Timestamp: time.Now(),
	}
	nonce, err := util.GenerateNonce(rand)
	if err != nil {
		return err
	}
	bs, err := util.SealJsonBox(reqData, nonce, s.serverID, s.clientKey)
	if err != nil {
		return err
	}

	req := deleteLicenseSessionReq{
		Data: bs,
		N:    nonce,
	}
	url := fmt.Sprintf("%s/%s", s.url, base64.URLEncoding.EncodeToString(s.clientID))
	return sendJsonRequest(ctx, http.MethodDelete, url, req, nil)
}
