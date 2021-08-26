package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type HTTPAdapter interface {
	Do(req *http.Request) (*http.Response, error)
}

type SASTClient struct {
	BaseURL string
	Adapter HTTPAdapter
	Token   *AccessToken
}

func NewSASTClient(baseURL string, adapter HTTPAdapter) (*SASTClient, error) {
	client := SASTClient{
		BaseURL: baseURL,
		Adapter: adapter,
	}
	return &client, nil
}

func (c *SASTClient) Authenticate(username, password string) error {
	req, err := CreateAccessTokenRequest(c.BaseURL, username, password)
	if err != nil {
		return err
	}

	resp, err := c.doRequest(req, http.StatusOK)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	c.Token = &AccessToken{}
	return json.Unmarshal(responseBody, c.Token)
}

func (c *SASTClient) GetUsersResponseBody() ([]byte, error) {
	req, err := CreateGetUsersRequest(c.BaseURL, c.Token)
	if err != nil {
		return []byte{}, err
	}

	resp, err := c.doRequest(req, http.StatusOK)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (c *SASTClient) GetTeamsResponseBody() ([]byte, error) {
	req, err := CreateGetTeamsRequest(c.BaseURL, c.Token)
	if err != nil {
		return []byte{}, err
	}

	resp, err := c.doRequest(req, http.StatusOK)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (c *SASTClient) doRequest(request *http.Request, expectStatusCode int) (*http.Response, error) {
	resp, err := c.Adapter.Do(request)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != expectStatusCode {
		return nil, fmt.Errorf("invalid response: %v", resp)
	}
	return resp, nil
}
