package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
)

const (
	UsersEndpoint = "/CxRestAPI/auth/Users"
	TeamsEndpoint = "/CxRestAPI/auth/Teams"
	RolesEndpoint = "/CxRestAPI/auth/Roles"

	LdapServersEndpoint            = "/CxRestAPI/auth/LDAPServers"
	LdapRoleMappingsEndpoint       = "/CxRestAPI/auth/LDAPRoleMappings"
	LdapTeamMappingsEndpoint       = "/CxRestAPI/auth/LDAPTeamMappings"
	SamlIdentityProvidersEndpoint  = "/CxRestAPI/auth/SamlIdentityProviders"
	SamlRoleMappingsEndpoint       = "/CxRestAPI/auth/SamlRoleMappings"
	TeamMappingsEndpoint           = "/CxRestAPI/auth/SamlTeamMappings"
	ReportsLastTriagedScanEndpoint = "/CxWebInterface/odata/v1/Results?$select=Id,ScanId,Date,Scan&$expand=Scan($select=ProjectId)&$filter="
	ReportsCheckStatusEndpoint     = "/CxRestAPI/help/reports/sastScan/%d/status"
	ReportsResultEndpoint          = "/CxRestAPI/help/reports/sastScan/%d"
	CreateReportIDEndpoint         = "/CxRestAPI/help/reports/sastScan"
	LastTriagedFilters             = "Date gt %s and Comment ne null"
)

type RetryableHTTPAdapter interface {
	Do(req *retryablehttp.Request) (*http.Response, error)
}

type SASTClient struct {
	BaseURL string
	Adapter RetryableHTTPAdapter
	Token   *AccessToken
}

type SASTError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func NewSASTClient(baseURL string, adapter RetryableHTTPAdapter) (*SASTClient, error) {
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

	resp, err := c.Adapter.Do(req)
	if err != nil {
		log.Debug().
			Err(err).
			Str("method", req.Method).
			Str("url", req.URL.String()).
			Msgf("authenticate failed request")
		return fmt.Errorf("authentication error - request failed")
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Debug().Err(closeErr).Msg("authenticate")
		}
	}()

	logger := log.With().
		Str("method", req.Method).
		Str("url", req.URL.String()).
		Int("statusCode", resp.StatusCode).
		Logger()

	if resp.StatusCode == http.StatusOK {
		responseBody, ioErr := ioutil.ReadAll(resp.Body)
		if ioErr != nil {
			logger.Debug().Err(ioErr).Msg("authenticate ok failed read response")
			return fmt.Errorf("authentication error - could not read response")
		}
		c.Token = &AccessToken{}
		unmarshalErr := json.Unmarshal(responseBody, c.Token)
		if unmarshalErr != nil {
			logger.Debug().
				Err(unmarshalErr).
				Str("responseBody", string(responseBody)).
				Msg("authenticate ok failed to unmarshal response")
			return fmt.Errorf("authentication error - could not decode response")
		}
		return nil
	} else if resp.StatusCode == http.StatusBadRequest {
		responseBody, ioErr := ioutil.ReadAll(resp.Body)
		if ioErr != nil {
			logger.Debug().Err(ioErr).Msg("authenticate bad request failed to read response")
			return fmt.Errorf("authentication error - could not read response")
		}
		var response SASTError
		unmarshalErr := json.Unmarshal(responseBody, &response)
		if unmarshalErr != nil {
			logger.Debug().
				Err(unmarshalErr).
				Str("responseBody", string(responseBody)).
				Msg("authenticate bad request failed to unmarshal response")
			return fmt.Errorf("authentication error - could not decode response")
		}
		if response.ErrorDescription == "invalid_username_or_password" {
			return fmt.Errorf("authentication error - please confirm your user name and password")
		}
	}

	logger.Debug().Msg("authenticate unexpected response")
	return fmt.Errorf("authentication error - please try again later or contact support")
}

func (c *SASTClient) GetResponseBody(endpoint string) ([]byte, error) {
	req, err := CreateRequest(http.MethodGet, c.BaseURL+endpoint, nil, c.Token)
	if err != nil {
		return []byte{}, err
	}

	resp, err := c.doRequest(req, http.StatusOK)
	if err != nil {
		return []byte{}, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Debug().Err(closeErr).Msg("getResponseBody")
		}
	}()
	return ioutil.ReadAll(resp.Body)
}

func (c *SASTClient) PostResponseBody(endpoint string, body io.Reader) ([]byte, error) {
	req, err := CreateRequest(http.MethodPost, c.BaseURL+endpoint, body, c.Token)
	if err != nil {
		return []byte{}, err
	}

	resp, err := c.doRequest(req, http.StatusAccepted)
	if err != nil {
		return []byte{}, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Debug().Err(closeErr).Msg("postResponseBody")
		}
	}()
	return ioutil.ReadAll(resp.Body)
}

func (c *SASTClient) doRequest(req *retryablehttp.Request, expectStatusCode int) (*http.Response, error) {
	resp, err := c.Adapter.Do(req)
	if err != nil {
		return nil, err
	}
	log.Debug().
		Err(err).
		Str("method", req.Method).
		Str("url", req.URL.String()).
		Int("statusCode", resp.StatusCode).
		Msg("request")
	if resp.StatusCode != expectStatusCode {
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Debug().Err(closeErr).Msg("doRequest")
			}
		}()
		return nil, fmt.Errorf("request %s %s failed with status code %d", req.Method, req.URL.String(), resp.StatusCode)
	}
	return resp, nil
}

func (c *SASTClient) GetReportStatusResponse(report ReportResponse) (*StatusResponse, error) {
	statusUnm, errGetStatus := c.GetReportIDStatus(report.ReportID)
	if errGetStatus != nil {
		return &StatusResponse{}, errGetStatus
	}

	var status StatusResponse
	errStatusSheriff := json.Unmarshal(statusUnm, &status)
	if errStatusSheriff != nil {
		return &StatusResponse{}, errStatusSheriff
	}

	return &status, nil
}

func (c *SASTClient) GetUsers() ([]byte, error) {
	return c.GetResponseBody(UsersEndpoint)
}

func (c *SASTClient) GetRoles() ([]byte, error) {
	return c.GetResponseBody(RolesEndpoint)
}

func (c *SASTClient) GetTeams() ([]byte, error) {
	return c.GetResponseBody(TeamsEndpoint)
}

func (c *SASTClient) GetLdapServers() ([]byte, error) {
	return c.GetResponseBody(LdapServersEndpoint)
}

func (c *SASTClient) GetLdapRoleMappings() ([]byte, error) {
	return c.GetResponseBody(LdapRoleMappingsEndpoint)
}

func (c *SASTClient) GetLdapTeamMappings() ([]byte, error) {
	return c.GetResponseBody(LdapTeamMappingsEndpoint)
}

func (c *SASTClient) GetSamlIdentityProviders() ([]byte, error) {
	return c.GetResponseBody(SamlIdentityProvidersEndpoint)
}

func (c *SASTClient) GetSamlRoleMappings() ([]byte, error) {
	return c.GetResponseBody(SamlRoleMappingsEndpoint)
}

func (c *SASTClient) GetSamlTeamMappings() ([]byte, error) {
	return c.GetResponseBody(TeamMappingsEndpoint)
}

func (c *SASTClient) GetTriagedScansFromDate(fromDate string, offset, limit int) ([]byte, error) {
	url := ReportsLastTriagedScanEndpoint
	url += GetEncodingURL(LastTriagedFilters, fromDate)
	url += fmt.Sprintf("&$skip=%d&$top=%d", offset, limit)
	return c.GetResponseBody(url)
}

func (c *SASTClient) GetReportIDStatus(reportID int) ([]byte, error) {
	return c.GetResponseBody(fmt.Sprintf(ReportsCheckStatusEndpoint, reportID))
}

func (c *SASTClient) GetReportResult(reportID int) ([]byte, error) {
	return c.GetResponseBody(fmt.Sprintf(ReportsResultEndpoint, reportID))
}

func (c *SASTClient) PostReportID(body io.Reader) ([]byte, error) {
	return c.PostResponseBody(CreateReportIDEndpoint, body)
}
