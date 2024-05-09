package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
)

type OidcClient struct {
	DatabaseFields
	ProviderPlatformID uint   `json:"provider_platform_id"`
	ClientID           string `gorm:"size:255" json:"client_id"`
	ClientName         string `gorm:"size:255" json:"client_name"`
	ClientSecret       string `gorm:"size:255" json:"client_secret"`
	RedirectURIs       string `gorm:"size:255" json:"redirect_uris"`
	Scopes             string `gorm:"size:255" json:"scopes"`

	Provider *ProviderPlatform `gorm:"foreignKey:ProviderPlatformID" json:"-"`
}

const DefaultScopes = "openid profile email"

func (OidcClient) TableName() string {
	return "oidc_clients"
}

func OidcClientFromProvider(prov *ProviderPlatform, autoRegister bool) (*OidcClient, string, error) {
	externalId := ""
	redirectURI := prov.GetDefaultRedirectURI()
	client := http.Client{}
	headers := map[string]string{}
	headers["Authorization"] = "Bearer " + os.Getenv("HYDRA_ADMIN_TOKEN")
	headers["Origin"] = os.Getenv("APP_URL")
	body := map[string]interface{}{}
	body["client_name"] = prov.Name
	body["redirect_uris"] = []string{redirectURI}
	body["scopes"] = DefaultScopes
	body["acces_token_strategy"] = "opaque"
	body["metadata"] = map[string]interface{}{
		"Origin": os.Getenv("APP_URL"),
	}
	body["allowed_cors_origins"] = []string{os.Getenv("HYDRA_ADMIN_URL"), os.Getenv("APP_URL"), os.Getenv("FRONTEND_URL"), os.Getenv("HYDRA_PUBLIC_URL")}
	body["grant_types"] = []string{"authorization_code"}
	body["authorization_code_grant_access_token_lifespan"] = "3h"
	body["authorization_code_grant_id_token_lifespan"] = "3h"
	body["authorization_code_grant_refresh_token_lifespan"] = "3h"
	body["skip_consent"] = false
	body["skip_logout_consent"] = true
	body["token_endpoint_auth_method"] = "client_secret_basic"
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, externalId, err
	}
	req, err := http.NewRequest("POST", os.Getenv("HYDRA_ADMIN_URL")+"/admin/clients", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, externalId, err
	}
	req.Header.Add("Authorization", headers["Authorization"])
	resp, err := client.Do(req)
	if err != nil {
		return nil, externalId, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, externalId, fmt.Errorf("error creating client: %s", resp.Status)
	}
	var clientData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&clientData)
	if err != nil {
		return nil, externalId, err
	}
	if clientData == nil {
		return nil, externalId, fmt.Errorf("error creating client: %s", resp.Status)
	}
	redirects := clientData["redirect_uris"].([]interface{})
	if len(redirects) == 0 {
		return nil, externalId, fmt.Errorf("error creating client: no redirect URIs")
	}
	receivedURI := redirects[0].(string)
	if receivedURI != redirectURI {
		return nil, externalId, fmt.Errorf("error creating client: redirect URI mismatch")
	}
	oidcClient := &OidcClient{
		ProviderPlatformID: prov.ID,
		ClientID:           clientData["client_id"].(string),
		ClientName:         clientData["client_name"].(string),
		ClientSecret:       clientData["client_secret"].(string),
		RedirectURIs:       receivedURI,
		Scopes:             clientData["scope"].(string),
	}
	if autoRegister && (prov.Type == CanvasCloud || prov.Type == CanvasOSS) {
		externalId, err = autoRegisterCanvas(prov, oidcClient)
		if err != nil {
			slog.Error("Error auto registering provider as client: ", err)
		}
	}
	return oidcClient, externalId, nil
}

func autoRegisterCanvas(prov *ProviderPlatform, oidcClient *OidcClient) (string, error) {
	client := http.Client{}
	externalId := ""
	/**** This is for canvas, kolibri will need another implementation *****/
	if prov.Type == CanvasCloud || prov.Type == CanvasOSS {
		// register the login client with canvas
		baseURL := prov.BaseUrl + "/api/v1/accounts/" + prov.AccountID + "/authentication_providers"
		// we need to add the client_id, client_secret, and redirect_uri to the request form encoded
		form := url.Values{}
		form.Add("auth_type", "openid_connect")
		form.Add("client_id", oidcClient.ClientID)
		form.Add("client_secret", oidcClient.ClientSecret)
		form.Add("authorize_url", os.Getenv("HYDRA_PUBLIC_URL")+"/oauth2/auth")
		form.Add("token_url", os.Getenv("HYDRA_PUBLIC_URL")+"/oauth2/auth")
		form.Add("userinfo_endpoint", os.Getenv("HYDRA_PUBLIC_URL")+"/userinfo")
		request, err := http.NewRequest("POST", baseURL, bytes.NewBufferString(form.Encode()))
		if err != nil {
			log.Println("Error creating request object: ", err)
			return "", err
		}
		log.Printf("Authorization: Bearer %s", prov.AccessKey)
		headers := make(map[string]string)
		headers["Content-Type"] = "application/x-www-form-urlencoded"
		headers["Authorization"] = "Bearer " + prov.AccessKey
		headers["Accept"] = "application/json"
		for k, v := range headers {
			request.Header.Add(k, v)
		}
		log.Printf("Request: %v", request)
		response, err := client.Do(request)
		if err != nil {
			log.Println("Error sending request: ", err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusCreated && response.StatusCode != http.StatusOK {
			log.Println("Error creating authentication provider: ", response.Status)
		}
		var authProvider map[string]interface{}
		err = json.NewDecoder(response.Body).Decode(&authProvider)
		if err != nil {
			slog.Error("Error decoding response body: ", err)
			return "", err
		}
		if authProvider == nil {
			slog.Error("Error creating authentication provider: ", response.Status, response.Body)
			return "", err
		}
		if authProvider["id"] == nil {
			slog.Error("Error creating authentication provider: no ID in response")
			return "", err
		}
		if id, ok := authProvider["id"].(int); ok {
			externalId = fmt.Sprintf("%d", id)
		}
	}
	return externalId, nil
}
