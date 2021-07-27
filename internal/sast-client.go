package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type SASTClient struct {
	BaseURL string
	Adapter *http.Client
	Token   *AccessToken
}

func NewSASTClient(baseURL string) (*SASTClient, error) {
	client := SASTClient{
		BaseURL: baseURL,
		Adapter: &http.Client{},
	}
	return &client, nil
}

func (c *SASTClient) Authenticate(username, password string) error {
	req, err := CreateAccessTokenRequest(c.BaseURL, username, password)
	if err != nil {
		return err
	}
	resp, err := c.Adapter.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid response: %v", resp)
	}
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	c.Token = &AccessToken{}
	return json.Unmarshal(responseBody, c.Token)
}

func (c *SASTClient) GetProjects() ([]Project, error) {
	var projects []Project
	req, err := CreateGetProjectsRequest(c.BaseURL, c.Token)
	if err != nil {
		return projects, err
	}
	resp, err := c.Adapter.Do(req)
	if err != nil {
		return projects, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return projects, fmt.Errorf("invalid response: %v", resp)
	}
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(responseBody, &projects)
	return projects, err
}
