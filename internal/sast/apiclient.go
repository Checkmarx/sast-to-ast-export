package sast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/checkmarxDev/ast-sast-export/internal/utils"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
)

const (
	usersEndpoint = "/CxRestAPI/auth/Users"
	teamsEndpoint = "/CxRestAPI/auth/Teams"
	rolesEndpoint = "/CxRestAPI/auth/Roles"

	ldapServersEndpoint           = "/CxRestAPI/auth/LDAPServers"
	ldapRoleMappingsEndpoint      = "/CxRestAPI/auth/LDAPRoleMappings"
	ldapTeamMappingsEndpoint      = "/CxRestAPI/auth/LDAPTeamMappings"
	samlIdentityProvidersEndpoint = "/CxRestAPI/auth/SamlIdentityProviders"
	samlRoleMappingsEndpoint      = "/CxRestAPI/auth/SamlRoleMappings"
	teamMappingsEndpoint          = "/CxRestAPI/auth/SamlTeamMappings"
	reportsCheckStatusEndpoint    = "/CxRestAPI/help/reports/sastScan/%d/status"
	reportsResultEndpoint         = "/CxRestAPI/help/reports/sastScan/%d"
	createReportIDEndpoint        = "/CxRestAPI/help/reports/sastScan"

	// ScanReportTypeXML defines SAST report type XML
	ScanReportTypeXML = "XML"
)

type Client interface {
	Authenticate(username, password string) error
	GetResponseBody(endpoint string) ([]byte, error)
	PostResponseBody(endpoint string, body io.Reader) ([]byte, error)
	GetUsers() ([]byte, error)
	GetRoles() ([]byte, error)
	GetTeams() ([]byte, error)
	GetLdapServers() ([]byte, error)
	GetLdapRoleMappings() ([]byte, error)
	GetLdapTeamMappings() ([]byte, error)
	GetSamlIdentityProviders() ([]byte, error)
	GetSamlRoleMappings() ([]byte, error)
	GetSamlTeamMappings() ([]byte, error)
	GetProjectsWithLastScanID(fromDate string, offset, limit int) (*[]ProjectWithLastScanID, error)
	GetTriagedResultsByScanID(scanID int) (*[]TriagedScanResult, error)
	CreateScanReport(scanID int, reportType string, retry utils.Retry) ([]byte, error)
}

type RetryableHTTPAdapter interface {
	Do(req *retryablehttp.Request) (*http.Response, error)
}

type APIClient struct {
	BaseURL string
	Adapter RetryableHTTPAdapter
	Token   *AccessToken
}

type APIError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func NewSASTClient(baseURL string, adapter RetryableHTTPAdapter) (*APIClient, error) {
	client := APIClient{
		BaseURL: baseURL,
		Adapter: adapter,
	}
	return &client, nil
}

func (c *APIClient) Authenticate(username, password string) error {
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
		return fmt.Errorf("authentication error - please confirm you can connect to SAST")
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
		var response APIError
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

func (c *APIClient) GetResponseBody(endpoint string) ([]byte, error) {
	req, err := CreateRequest(http.MethodGet, c.BaseURL+endpoint, nil, c.Token)
	if err != nil {
		return []byte{}, err
	}
	return c.getResponseBodyFromRequest(req)
}

func (c *APIClient) getResponseBodyFromRequest(req *retryablehttp.Request) ([]byte, error) {
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

func (c *APIClient) PostResponseBody(endpoint string, body io.Reader) ([]byte, error) {
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

func (c *APIClient) doRequest(req *retryablehttp.Request, expectStatusCode int) (*http.Response, error) {
	resp, err := c.Adapter.Do(req)
	if err != nil {
		return nil, err
	}
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

func (c *APIClient) getReportStatusResponse(report ReportResponse) (*StatusResponse, error) {
	statusUnm, errGetStatus := c.getReportIDStatus(report.ReportID)
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

func (c *APIClient) GetUsers() ([]byte, error) {
	return c.GetResponseBody(usersEndpoint)
}

func (c *APIClient) GetRoles() ([]byte, error) {
	return c.GetResponseBody(rolesEndpoint)
}

func (c *APIClient) GetTeams() ([]byte, error) {
	return c.GetResponseBody(teamsEndpoint)
}

func (c *APIClient) GetLdapServers() ([]byte, error) {
	return c.GetResponseBody(ldapServersEndpoint)
}

func (c *APIClient) GetLdapRoleMappings() ([]byte, error) {
	return c.GetResponseBody(ldapRoleMappingsEndpoint)
}

func (c *APIClient) GetLdapTeamMappings() ([]byte, error) {
	return c.GetResponseBody(ldapTeamMappingsEndpoint)
}

func (c *APIClient) GetSamlIdentityProviders() ([]byte, error) {
	return c.GetResponseBody(samlIdentityProvidersEndpoint)
}

func (c *APIClient) GetSamlRoleMappings() ([]byte, error) {
	return c.GetResponseBody(samlRoleMappingsEndpoint)
}

func (c *APIClient) GetSamlTeamMappings() ([]byte, error) {
	return c.GetResponseBody(teamMappingsEndpoint)
}

func (c *APIClient) getReportIDStatus(reportID int) ([]byte, error) {
	return c.GetResponseBody(fmt.Sprintf(reportsCheckStatusEndpoint, reportID))
}

func (c *APIClient) getReportResult(reportID int) ([]byte, error) {
	return c.GetResponseBody(fmt.Sprintf(reportsResultEndpoint, reportID))
}

func (c *APIClient) postReportID(body io.Reader) ([]byte, error) {
	return c.PostResponseBody(createReportIDEndpoint, body)
}

func (c *APIClient) GetProjectsWithLastScanID(fromDate string, offset, limit int) (*[]ProjectWithLastScanID, error) {
	url := fmt.Sprintf("%s/Cxwebinterface/odata/v1/Projects", c.BaseURL)
	req, requestErr := CreateRequest(http.MethodGet, url, nil, c.Token)
	if requestErr != nil {
		return nil, requestErr
	}
	q := req.URL.Query()
	q.Add("$select", "Id,LastScanId")
	q.Add("$expand", "LastScan($select=Id)")
	q.Add("$filter", fmt.Sprintf("LastScan/ScanCompletedOn gt %s", fromDate))
	q.Add("$skip", fmt.Sprintf("%d", offset))
	q.Add("$top", fmt.Sprintf("%d", limit))
	req.URL.RawQuery = q.Encode()
	body, getErr := c.getResponseBodyFromRequest(req)
	if getErr != nil {
		return nil, getErr
	}
	var response ODataProjectsWithLastScanID
	unmarshalErr := json.Unmarshal(body, &response)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	return &response.Value, nil
}

func (c *APIClient) GetTriagedResultsByScanID(scanID int) (*[]TriagedScanResult, error) {
	url := fmt.Sprintf("%s/Cxwebinterface/odata/v1/Scans(%d)/Results", c.BaseURL, scanID)
	req, requestErr := CreateRequest(http.MethodGet, url, nil, c.Token)
	if requestErr != nil {
		return nil, requestErr
	}
	q := req.URL.Query()
	q.Add("$filter", "Comment ne null")
	req.URL.RawQuery = q.Encode()
	body, getErr := c.getResponseBodyFromRequest(req)
	if getErr != nil {
		return nil, getErr
	}
	var response ODataTriagedResultsByScan
	unmarshalErr := json.Unmarshal(body, &response)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	return &response.Value, nil
}

func (c *APIClient) CreateScanReport(scanID int, reportType string, retry utils.Retry) ([]byte, error) {
	reportBody := &ReportRequest{
		ReportType: reportType,
		ScanID:     scanID,
	}
	reportJSON, marshalErr := json.Marshal(reportBody)
	if marshalErr != nil {
		return []byte{}, marshalErr
	}
	body := bytes.NewBuffer(reportJSON)
	log.Debug().
		Int("scanID", scanID).
		Str("type", reportType).
		Msg("creating report")
	postResponse, createErr := c.postReportID(body)
	if createErr != nil {
		return []byte{}, createErr
	}
	var reportCreateResponse ReportResponse
	unmarshalErr := json.Unmarshal(postResponse, &reportCreateResponse)
	if unmarshalErr != nil {
		return []byte{}, unmarshalErr
	}
	for i := 1; i <= retry.Attempts; i++ {
		time.Sleep(retryablehttp.DefaultBackoff(retry.MinSleep, retry.MaxSleep, i, nil))
		log.Debug().
			Int("attempt", i).
			Int("scanID", scanID).
			Str("type", reportType).
			Msg("getting report")
		status, statusFetchErr := c.getReportStatusResponse(reportCreateResponse)
		if statusFetchErr != nil {
			return []byte{}, statusFetchErr
		}
		if status.Status.Value == "Created" {
			reportData, getReportErr := c.getReportResult(reportCreateResponse.ReportID)
			if getReportErr != nil {
				return []byte{}, getReportErr
			}
			return reportData, nil
		}
	}
	return []byte{}, fmt.Errorf("failed getting report after %d attempts", retry.Attempts)
}
