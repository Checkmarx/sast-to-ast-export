package sast

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	accept                    = "Accept"
	contentType               = "Content-Type"
	authorization             = "Authorization"
	jsonContentType           = "application/json;v=1.0"
	formURLEncodedContentType = "application/x-www-form-urlencoded"
)

func CreateAccessTokenRequest(baseURL, username, password string) (*retryablehttp.Request, error) {
	tokenURL := fmt.Sprintf("%s/CxRestAPI/auth/identity/connect/token", baseURL)
	data := url.Values{}
	data.Add("username", username)
	data.Add("password", password)
	data.Add("grant_type", "password")
	data.Add("scope", "access_control_api sast_api")
	data.Add("client_id", "resource_owner_sast_client")
	data.Add("client_secret", "014DF517-39D1-4453-B7B3-9930C563627C")
	req, err := retryablehttp.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add(contentType, formURLEncodedContentType)
	req.Header.Add(accept, formURLEncodedContentType)
	return req, nil
}

func CreateRequest(httpMethod, requestURL string, requestBody io.Reader, token *AccessToken) (*retryablehttp.Request, error) {
	resp, err := retryablehttp.NewRequest(httpMethod, requestURL, requestBody)
	if err != nil {
		return nil, err
	}

	resp.Header.Add(contentType, jsonContentType)
	if token != nil {
		resp.Header.Add(authorization, fmt.Sprintf("%s %s", token.TokenType, token.AccessToken))
	}
	return resp, nil
}
