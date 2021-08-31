package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	JSONContentType = "application/json"
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

func CreateRequest(httpMethod, url string, requestBody io.Reader, token *AccessToken) (*http.Request, error) {
	resp, err := http.NewRequest(httpMethod, url, requestBody)
	if err != nil {
		return nil, err
	}

	resp.Header.Add("Content-Type", JSONContentType)
	resp.Header.Add("Authorization", fmt.Sprintf("%s %s", token.TokenType, token.AccessToken))
	return resp, nil
}

func dataToJSONReader(data interface{}) io.Reader {
	jsonStr, err := json.Marshal(data)
	if err != nil {
		fmt.Errorf("failed to stringify request data: %s", err)
	}
	return bytes.NewBuffer(jsonStr)
}

func CreateGetUsersRequest(baseURL string, token *AccessToken) (*http.Request, error) {
	return CreateRequest(http.MethodGet, fmt.Sprintf("%s/CxRestAPI/auth/Users", baseURL), nil, token)
}

func CreateGetTeamsRequest(baseURL string, token *AccessToken) (*http.Request, error) {
	return CreateRequest(http.MethodGet, fmt.Sprintf("%s/CxRestAPI/auth/Teams", baseURL), nil, token)
}

func GetListProjectsRequest(baseURL string, token *AccessToken) (*http.Request, error) {
	return CreateRequest(http.MethodGet, fmt.Sprintf("%s/CxRestAPI/help/projects", baseURL), nil, token)
}

func GetLastScanDataRequest(baseURL string, projectId int, token *AccessToken) (*http.Request, error) {
	return CreateRequest(http.MethodGet, fmt.Sprintf("%s/CxRestAPI/help/sast/scans?ProjectId=%d&last=1", baseURL, projectId), nil, token)
}

func GetReportIDStatusRequest(baseURL, reportId string, token *AccessToken) (*http.Request, error) {
	return CreateRequest(http.MethodGet, fmt.Sprintf("%s/CxRestAPI/help/reports/sastScan/%s/status", baseURL, reportId), nil, token)
}
