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

func createProduct(c *core.Core) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "create product"
		licenseIssuerID, err := strconv.Atoi(mux.Vars(r)["LICENSE_ISSUER_ID"])
		if err != nil {
			return responseBadRequestf("license issuer id: %v", err)
		}

		req := model.Product{
			Active:   true,
			IssuerID: licenseIssuerID,
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

		p, err := c.NewProduct(r.Context(), li, &req)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrInvalidInput):
				return responseBadRequest(err)
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}
		return responseJson(http.StatusCreated, p)
	}
}

func getAllProducts(c *core.Core) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "get all products"
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

		pp, err := c.GetAllProductsByIssuer(r.Context(), licenseIssuerID)
		if err != nil {
			logError(err, scope)
			return responseInternalServerError()
		}
		if pp == nil {
			pp = make([]*model.Product, 0)
		}
		return responseJson(http.StatusOK, pp)
	}
}

func getProduct(c *core.Core) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "get product"
		vars := mux.Vars(r)
		licenseIssuerID, err := strconv.Atoi(vars["LICENSE_ISSUER_ID"])
		if err != nil {
			return responseBadRequestf("license issuer id: %v", err)
		}
		productID, err := strconv.Atoi(vars["PRODUCT_ID"])
		if err != nil {
			return responseBadRequestf("product id: %v", err)
		}

		p, err := c.GetProduct(r.Context(), productID)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound()
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}
		if licenseIssuerID != p.IssuerID {
			return responseNotFound()
		}
		return responseJson(http.StatusOK, p)
	}
}

func updateProduct(c *core.Core) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "update product"
		vars := mux.Vars(r)
		licenseIssuerID, err := strconv.Atoi(vars["LICENSE_ISSUER_ID"])
		if err != nil {
			return responseBadRequestf("license issuer id: %v", err)
		}
		productID, err := strconv.Atoi(vars["PRODUCT_ID"])
		if err != nil {
			return responseBadRequestf("product id: %v", err)
		}

		data, err := readAllLim(r.Body)
		if err != nil {
			return responseBadRequest(err)
		}
		p := &model.Product{
			ID:       productID,
			IssuerID: licenseIssuerID,
		}
		err = json.Unmarshal(data, p)
		if err != nil {
			return responseBadRequest(err)
		}

		changes, err := core.UnmarshalChanges(data)
		if err != nil {
			return responseBadRequest(err) // should never happen
		}
		mask, _ := c.AuthorizeProductUpdate(login)
		field, ok := core.ChangesInMask(changes, mask)
		if !ok {
			return responseBadRequestf("unauthorized to change field: %s", field)
		}

		err = c.UpdateProduct(r.Context(), p, changes)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrInvalidInput):
				return responseBadRequest(err)
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}

		p, err = c.GetProduct(r.Context(), productID)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrNotFound):
				return responseNotFound()
			default:
				logError(err, scope)
				return responseInternalServerError()
			}
		}
		return responseJson(http.StatusOK, p)
	}
}

func deleteProduct(c *core.Core) apiAuthHandler {
	return func(r *http.Request, login *model.LicenseIssuer) *apiResponse {
		const scope = "delete product"
		vars := mux.Vars(r)
		licenseIssuerID, err := strconv.Atoi(vars["LICENSE_ISSUER_ID"])
		if err != nil {
			return responseBadRequestf("license issuer id: %v", err)
		}
		productID, err := strconv.Atoi(vars["PRODUCT_ID"])
		if err != nil {
			return responseBadRequestf("product id: %v", err)
		}

		_, canDelete := c.AuthorizeProductUpdate(login)
		if !canDelete {
			return responseForbidden()
		}
		err = c.DeleteProduct(r.Context(), productID, licenseIssuerID)
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
