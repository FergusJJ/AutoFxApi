package whop

import (
	"api/config"
	"encoding/json"
	"fmt"
	"log"

	"github.com/valyala/fasthttp"
)

func CheckLicenseKey(lisenceKey string) (string, string, error) {
	requestUri := fmt.Sprintf("https://api.whop.com/api/v2/memberships/%s/validate_license", lisenceKey)
	validateLicenseApiKey, err := config.Config("WHOP_VALIDATE_LICENSE_API_KEY")
	if err != nil {
		log.Println(err)
		return "", "", err
	}
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(requestUri)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validateLicenseApiKey))
	req.AppendBodyString("{\"metadata\":{\"newKey\":\"New Value\"}}")
	resp := fasthttp.AcquireResponse()
	if err := fasthttp.Do(req, resp); err != nil {
		return "", "", err
	}
	jsonResp := WhopValidateLicenseResponse{}
	if err := json.Unmarshal(resp.Body(), &jsonResp); err != nil {
		return "", "", err
	}

	if jsonResp.ID == "" {
		//404 - No such Membership found with the provided ID: validate_license
		jsonErrResp := WhopErrorResponse{}
		err := json.Unmarshal(resp.Body(), &jsonErrResp)
		if err != nil {
			return "", "", err
		}
		err = fmt.Errorf("%d|%s", jsonErrResp.Error.Status, jsonErrResp.Error.Message)
		return "", "", err
	}

	return jsonResp.ID, jsonResp.Email, nil
}
