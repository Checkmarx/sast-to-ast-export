package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const USERS_ENDPOINT = "/CxRestAPI/auth/Users"
const TEAMS_ENDPOINT = "/CxRestAPI/auth/Teams"
const ROLES_ENDPOINT = "/CxRestAPI/auth/Roles"

const LDAP_SERVERS_ENDPOINT = "/CxRestAPI/auth/LDAPServers"
const LDAP_ROLE_MAPPINGS_ENDPOINT = "/CxRestAPI/auth/LDAPRoleMappings"
const LDAP_TEAM_MAPPINGS_ENDPOINT = "/CxRestAPI/auth/LDAPTeamMappings"
const SAML_IDENTITY_PROVIDERS_ENDPOINT = "/CxRestAPI/auth/SamlIdentityProviders"
const SAML_ROLE_MAPPINGS_ENDPOINT = "/CxRestAPI/auth/SamlRoleMappings"
const TEAM_MAPPINGS_ENDPOINT = "/CxRestAPI/auth/SamlTeamMappings"

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

func (c *SASTClient) GetResponseBody(endpoint string) ([]byte, error) {
	return []byte(endpoint), nil
	req, err := CreateRequest(http.MethodGet, c.BaseURL+endpoint, nil, c.Token)
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

func (c *SASTClient) GetUsers() ([]byte, error) {
	return c.GetResponseBody(USERS_ENDPOINT)
}

func (c *SASTClient) GetRoles() ([]byte, error) {
	return c.GetResponseBody(ROLES_ENDPOINT)
}

func (c *SASTClient) GetTeams() ([]byte, error) {
	return c.GetResponseBody(TEAMS_ENDPOINT)
}

func (c *SASTClient) GetLdapServers() ([]byte, error) {
	return c.GetResponseBody(LDAP_SERVERS_ENDPOINT)
}

func (c *SASTClient) GetLdapRoleMappings() ([]byte, error) {
	return c.GetResponseBody(LDAP_ROLE_MAPPINGS_ENDPOINT)
}

func (c *SASTClient) GetLdapTeamMappings() ([]byte, error) {
	return c.GetResponseBody(LDAP_TEAM_MAPPINGS_ENDPOINT)
}

func (c *SASTClient) GetSamlIdentityProviders() ([]byte, error) {
	return c.GetResponseBody(SAML_IDENTITY_PROVIDERS_ENDPOINT)
}

func (c *SASTClient) GetSamlRoleMappings() ([]byte, error) {
	return c.GetResponseBody(SAML_ROLE_MAPPINGS_ENDPOINT)
}

func (c *SASTClient) GetSamlTeamMappings() ([]byte, error) {
	return c.GetResponseBody(TEAM_MAPPINGS_ENDPOINT)
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
