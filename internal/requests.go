package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func CreateAccessTokenRequest(baseURL, username, password string) (*http.Request, error) {
	tokenURL := fmt.Sprintf("%s/CxRestAPI/auth/identity/connect/token", baseURL)
	data := url.Values{}
	data.Add("username", username)
	data.Add("password", password)
	data.Add("grant_type", "password")
	data.Add("scope", "sast_rest_api access_control_api")
	data.Add("client_id", "resource_owner_client")
	data.Add("client_secret", "014DF517-39D1-4453-B7B3-9930C563627C")
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/x-www-form-urlencoded")
	return req, nil
}

func CreateRequest(httpMethod, url string, body interface{}, token *AccessToken) (*http.Request, error) {
	requestByte, errConvert := json.Marshal(body)
	if errConvert != nil {
		return nil, errConvert
	}

	resp, err := http.NewRequest(httpMethod, url, bytes.NewReader(requestByte))
	if err != nil {
		return nil, err
	}
	resp.Header.Add("Authorization", fmt.Sprintf("%s %s", token.TokenType, token.AccessToken))
	return resp, nil
}
