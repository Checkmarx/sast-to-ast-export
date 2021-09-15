package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	"github.com/rs/zerolog/log"
)

var (
	exportData []ExportData
	usersData,
	rolesData,
	ldapRolesData,
	samlRolesData,
	teamsData,
	samlTeamsData,
	ldapTeamsData,
	samlIDpsData,
	ldapServersData []byte
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

var isDebug bool

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

func (c *SASTClient) PostResponseBody(endpoint string, body io.Reader) ([]byte, error) {
	req, err := CreateRequest(http.MethodPost, c.BaseURL+endpoint, body, c.Token)
	if err != nil {
		panic(err)
	}

	resp, err := c.doRequest(req, http.StatusAccepted)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (c *SASTClient) doRequest(request *http.Request, expectStatusCode int) (*http.Response, error) {
	resp, err := c.Adapter.Do(request)
	if err != nil {
		log.Debug().
			Str("method", request.Method).
			Str("url", request.URL.String()).
			Msg("request")
		return nil, err
	}
	if resp.StatusCode != expectStatusCode {
		defer resp.Body.Close()
		responseContent, dumpErr := httputil.DumpResponse(resp, true)
		if dumpErr != nil {
			log.Debug().
				Str("error", dumpErr.Error()).
				Msg("failed dumping unexpected response")
			log.Debug().
				Str("method", request.Method).
				Str("url", request.URL.String()).
				Int("statusCode", resp.StatusCode).
				Msg("request")
		} else {
			log.Debug().
				Str("method", request.Method).
				Str("url", request.URL.String()).
				Int("statusCode", resp.StatusCode).
				Str("content", string(responseContent)).
				Msg("request")
		}
		return nil, fmt.Errorf("request %s %s failed with status code %d", request.Method, request.URL.String(), resp.StatusCode)
	}
	log.Debug().
		Str("method", request.Method).
		Str("url", request.URL.String()).
		Int("statusCode", resp.StatusCode).
		Msg("request")
	return resp, nil
}

func GetReportStatusResponse(c *SASTClient, report ReportResponse) (*StatusResponse, error) {
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

func (c *SASTClient) GetLastTriagedScanData(fromDate string) ([]byte, error) {
	return c.GetResponseBody(ReportsLastTriagedScanEndpoint + GetEncodingURL(LastTriagedFilters, fromDate))
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
