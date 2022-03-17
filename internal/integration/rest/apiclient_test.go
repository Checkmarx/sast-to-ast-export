package rest

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
)

const (
	BaseURL                 = "http://127.0.0.1"
	ErrorResponseJSON       = `"error"`
	InvalidDataResponseJSON = `invalid data`
)

var (
	mockToken = &AccessToken{AccessToken: "jwt", TokenType: "Bearer", ExpiresIn: 1234}
)

type DoResponse struct {
	Response *http.Response
	Err      error
}

type HTTPClientMock struct {
	DoResponse *http.Response
	DoError    error
}

func (c *HTTPClientMock) Do(_ *retryablehttp.Request) (*http.Response, error) {
	return c.DoResponse, c.DoError
}

type HTTPClientMock2 struct {
	DoHandler func(*retryablehttp.Request) (*http.Response, error)
}

func (c *HTTPClientMock2) Do(request *retryablehttp.Request) (*http.Response, error) {
	return c.DoHandler(request)
}

func newMockClient(response *http.Response) (*APIClient, error) {
	adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
	client, err := NewSASTClient(BaseURL, adapter)
	if err != nil {
		return nil, err
	}
	client.Token = mockToken
	return client, nil
}

func TestNewSASTClient(t *testing.T) {
	response := http.Response{
		Body: io.NopCloser(bytes.NewBufferString("test")),
	}
	adapter := &HTTPClientMock{DoResponse: &response, DoError: nil}

	client, err := NewSASTClient(BaseURL, adapter)

	assert.NoError(t, err)
	assert.Equal(t, BaseURL, client.BaseURL)
	assert.Equal(t, adapter, client.Adapter)
}

func TestAPIClient_Authenticate(t *testing.T) {
	t.Run("authenticates successfully", func(t *testing.T) {
		responseJSON := `{"access_token":"jwt", "token_type":"Bearer", "expires_in": 1234}`
		response := makeOkResponse(responseJSON) //nolint:bodyclose
		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
		client, _ := NewSASTClient(BaseURL, adapter)

		err := client.Authenticate("username", "password")

		assert.NoError(t, err)
		assert.NotNil(t, client.Token)
		expected := &AccessToken{AccessToken: "jwt", TokenType: "Bearer", ExpiresIn: 1234}
		assert.Equal(t, client.Token, expected)
	})
	t.Run("returns error if response is not HTTP OK", func(t *testing.T) {
		response := makeBadRequestResponse(ErrorResponseJSON)
		defer response.Body.Close()
		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
		client, _ := NewSASTClient(BaseURL, adapter)

		err := client.Authenticate("username", "password")

		assert.Error(t, err)
		assert.Nil(t, client.Token)
	})
	t.Run("returns error if can't parse response", func(t *testing.T) {
		response := makeOkResponse(InvalidDataResponseJSON) //nolint:bodyclose
		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
		client, _ := NewSASTClient(BaseURL, adapter)

		err := client.Authenticate("username", "password")

		assert.Error(t, err)
		assert.Equal(t, &AccessToken{}, client.Token)
	})
	t.Run("fails if can't connect to server", func(t *testing.T) {
		response := http.Response{
			StatusCode: 0,
			Status:     "Unknown",
			Body:       io.NopCloser(bytes.NewBufferString("")),
		}
		adapter := &HTTPClientMock{DoResponse: &response, DoError: fmt.Errorf("can't connect to server")}
		client, _ := NewSASTClient(BaseURL, adapter)

		err := client.Authenticate("username", "password")

		assert.EqualError(t, err, "authentication error - please confirm you can connect to SAST")
	})
	t.Run("fails if credentials are incorrect", func(t *testing.T) {
		responseJSON := `{"error": "invalid_grant","error_description": "invalid_username_or_password"}`
		response := makeBadRequestResponse(responseJSON) //nolint:bodyclose
		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
		client, _ := NewSASTClient(BaseURL, adapter)

		err := client.Authenticate("username", "password")

		assert.EqualError(t, err, "authentication error - please confirm your user name and password")
	})
}

func TestAPIClient_doRequest(t *testing.T) {
	t.Run("returns successful response", func(t *testing.T) {
		request, err := retryablehttp.NewRequest("GET", "http://localhost/test", nil)
		assert.NoError(t, err)
		expectedStatusCode := 200
		responseJSON := `{"data": "some data"}`
		adapter := &HTTPClientMock{DoResponse: makeOkResponse(responseJSON), DoError: nil} //nolint:bodyclose
		client, _ := NewSASTClient(BaseURL, adapter)
		client.Token = mockToken

		result, err := client.doRequest(request, expectedStatusCode)
		defer func() {
			closeErr := result.Body.Close()
			assert.NoError(t, closeErr)
		}()
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, result.StatusCode, expectedStatusCode)

		content, ioErr := io.ReadAll(result.Body)
		assert.NoError(t, ioErr)
		assert.Equal(t, responseJSON, string(content))
	})

	t.Run("returns error if response is not the expected one", func(t *testing.T) {
		request, err := retryablehttp.NewRequest("GET", "http://localhost/test", nil)
		assert.NoError(t, err)
		expectedStatusCode := 400
		adapter := &HTTPClientMock{DoResponse: makeBadRequestResponse(ErrorResponseJSON), DoError: nil} //nolint:bodyclose
		client, _ := NewSASTClient(BaseURL, adapter)
		client.Token = mockToken

		result, err := client.doRequest(request, expectedStatusCode)
		defer func() {
			closeErr := result.Body.Close()
			assert.NoError(t, closeErr)
		}()
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, result.StatusCode, expectedStatusCode)
	})
}

// nolint:dupl
func TestAPIClient_GetUsers(t *testing.T) {
	t.Run("returns users response", func(t *testing.T) {
		responseJSON := `[{"id": 1, "userName": "test1", "lastLoginDate": "2021-08-17T12:22:28.2331383Z", "active": true},
						  {"id": 2, "userName": "test2", "lastLoginDate": "2021-08-17T12:22:28.2331383Z", "active": true},
						  {"id": 3, "userName": "test3", "lastLoginDate": "2021-08-17T12:22:28.2331383Z", "active": true}]`
		client, clientErr := newMockClient(makeOkResponse(responseJSON)) //nolint:bodyclose
		assert.NoError(t, clientErr)

		result, err := client.GetUsers()

		assert.NoError(t, err)
		expected := []*User{
			{ID: 1, UserName: "test1", LastLoginDate: "2021-08-17T12:22:28.2331383Z", Active: true},
			{ID: 2, UserName: "test2", LastLoginDate: "2021-08-17T12:22:28.2331383Z", Active: true},
			{ID: 3, UserName: "test3", LastLoginDate: "2021-08-17T12:22:28.2331383Z", Active: true},
		}
		assert.Equal(t, expected, result)
	})
	t.Run("returns error if response is not HTTP OK", func(t *testing.T) {
		client, clientErr := newMockClient(makeBadRequestResponse(ErrorResponseJSON)) //nolint:bodyclose
		assert.NoError(t, clientErr)

		result, err := client.GetUsers()

		assert.Error(t, err)
		assert.Len(t, result, 0)
	})
}

// nolint:dupl
func TestAPIClient_GetTeams(t *testing.T) {
	t.Run("returns teams response", func(t *testing.T) {
		responseJSON := `[{"id": 1, "name": "test1", "fullName": "/CxServer/test1", "parentId": 0},
						  {"id": 2, "name": "test2", "fullName": "/CxServer/test2", "parentId": 1},
						  {"id": 3, "name": "test3", "fullName": "/CxServer/test3", "parentId": 1}]`
		client, clientErr := newMockClient(makeOkResponse(responseJSON)) //nolint:bodyclose
		assert.NoError(t, clientErr)

		result, err := client.GetTeams()

		assert.NoError(t, err)
		expected := []*Team{
			{ID: 1, Name: "test1", FullName: "/CxServer/test1", ParendID: 0},
			{ID: 2, Name: "test2", FullName: "/CxServer/test2", ParendID: 1},
			{ID: 3, Name: "test3", FullName: "/CxServer/test3", ParendID: 1},
		}
		assert.Equal(t, expected, result)
	})
	t.Run("returns error if response is not HTTP OK", func(t *testing.T) {
		client, clientErr := newMockClient(makeBadRequestResponse(ErrorResponseJSON)) //nolint:bodyclose
		assert.NoError(t, clientErr)

		result, err := client.GetTeams()

		assert.Error(t, err)
		assert.Len(t, result, 0)
	})
}

func TestAPIClient_GetSamlTeamMappings(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		responseJSON := `[{"id":3,"samlIdentityProviderId":2,"teamId":4,"teamFullPath":"/CxServer/Mapped SAML","samlAttributeValue":"TeamA"}]`
		client, clientErr := newMockClient(makeOkResponse(responseJSON)) //nolint:bodyclose
		assert.NoError(t, clientErr)

		result, err := client.GetSamlTeamMappings()

		assert.NoError(t, err)
		expected := []*SamlTeamMapping{
			{ID: 3, SamlIdentityProviderID: 2, TeamID: 4, TeamFullPath: "/CxServer/Mapped SAML", SamlAttributeValue: "TeamA"},
		}
		assert.Equal(t, expected, result)
	})
	t.Run("failure case", func(t *testing.T) {
		client, clientErr := newMockClient(makeBadRequestResponse(ErrorResponseJSON)) //nolint:bodyclose
		assert.NoError(t, clientErr)

		result, err := client.GetSamlTeamMappings()

		assert.Error(t, err)
		assert.Len(t, result, 0)
	})
}

// nolint:dupl
func TestAPIClient_GetRoles(t *testing.T) {
	t.Run("returns teams response", func(t *testing.T) {
		responseJSON := `[{"id": 1, "isSystemRole": true, "name": "test1", "description": "test1", permissionIds: []},
						  {"id": 2, "isSystemRole": true, "name": "test2", "description": "test2", permissionIds: []},
						  {"id": 3, "isSystemRole": true, "name": "test3", "description": "test3", permissionIds: []}]`
		response := makeOkResponse(responseJSON) //nolint:bodyclose
		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
		client, _ := NewSASTClient(BaseURL, adapter)
		client.Token = mockToken

		result, err := client.GetRoles()

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, responseJSON, string(result))
	})
	t.Run("returns error if response is not HTTP OK", func(t *testing.T) {
		response := makeBadRequestResponse(ErrorResponseJSON)
		defer response.Body.Close()
		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
		client, _ := NewSASTClient(BaseURL, adapter)
		client.Token = mockToken

		result, err := client.GetRoles()

		assert.Error(t, err)
		assert.Len(t, result, 0)
	})
}

func TestAPIClient_GetProjectsWithLastScanID(t *testing.T) {
	odataResponse := `
{
    "@odata.context": "http://localhost/odata/$metadata#somecontext",
    "value": [
        {"Id": 1,"LastScanId": 1000000,"LastScan": {"Id": 1000000}},
        {"Id": 2,"LastScanId": 1000001,"LastScan": {"Id": 1000001}}
	]
}`
	adapter := &HTTPClientMock{DoResponse: makeOkResponse(odataResponse), DoError: nil} //nolint:bodyclose
	client, _ := NewSASTClient(BaseURL, adapter)
	client.Token = mockToken

	result, err := client.GetProjectsWithLastScanID("2021-10-7", 0, 10)

	expected := []ProjectWithLastScanID{
		{ID: 1, LastScanID: 1000000},
		{ID: 2, LastScanID: 1000001},
	}
	assert.NoError(t, err)
	assert.Equal(t, expected, *result)
}

func TestAPIClient_GetTriagedResultsByScanID(t *testing.T) {
	odataResponse := `
{
    "@odata.context": "http://localhost/odata/$metadata#somecontext",
    "value": [
        {"Id": 2},
        {"Id": 3}
	]
}`
	response := makeOkResponse(odataResponse)
	defer response.Body.Close()
	adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
	client, _ := NewSASTClient(BaseURL, adapter)
	client.Token = mockToken

	result, err := client.GetTriagedResultsByScanID(1000000)

	expected := []TriagedScanResult{
		{ID: 2},
		{ID: 3},
	}
	assert.NoError(t, err)
	assert.Equal(t, expected, *result)
}

// nolint:funlen
func TestAPIClient_CreateScanReport(t *testing.T) {
	retry := Retry{Attempts: 3, MinSleep: time.Millisecond, MaxSleep: time.Millisecond}
	t.Run("success case", func(t *testing.T) {
		scanID := 1000000
		// nolint:goconst
		postReportURL := "/CxRestAPI/help/reports/sastScan"
		getReportStatusURL := "/CxRestAPI/help/reports/sastScan/1250/status"
		getReportURL := "/CxRestAPI/help/reports/sastScan/1250"
		reportCreateJSON := `{
					"reportId": 1250, 
					"links": {
						"report": {"rel": "content", "uri": "/reports/sastScan/1250"},
						"status": {"rel": "status","uri": "/reports/sastScan/1250/status"}
					}
				}`
		reportStatusJSON1 := `{
					"link": {"rel": "content","uri": "/reports/sastScan/1250"},
					"contentType": "application/xml",
					"status": {"id": 2,"value": "In progress"}
				}`
		// nolint:goconst
		reportStatusJSON2 := `{
					"link": {"rel": "content","uri": "/reports/sastScan/1250"},
					"contentType": "application/xml",
					"status": {"id": 2,"value": "Created"}
				}`
		reportXML := `<?xml version="1.0" encoding="utf-8"?>
<CxXMLResults Owner="admin" ScanId="1000000" ProjectId="1" ...>
    <Query id="600" ...>
        <Result NodeId="10000000004" ...>
            <Path ResultId="1000000" ...>
                <PathNode><FileName>path/to/file.java</FileName><Line>13</Line>...</PathNode>
            </Path>
        </Result>
        <Result NodeId="10000000005" ...>
            <Path ResultId="1000000" ...>
                <PathNode><FileName>path/to/file2.java</FileName><Line>32</Line>...</PathNode>
            </Path>
        </Result>
		...
    </Query>
	...
</CxXMLResults>`
		responses := map[string][]DoResponse{
			postReportURL: {
				{Response: makeResponse(202, "Created", reportCreateJSON)}, //nolint:bodyclose
			},
			getReportStatusURL: {
				{Response: makeOkResponse(reportStatusJSON1)}, //nolint:bodyclose
				{Response: makeOkResponse(reportStatusJSON2)}, //nolint:bodyclose
			},
			getReportURL: {
				{Response: makeOkResponse(reportXML)}, //nolint:bodyclose
			},
		}
		client := makeCreateReportClient(responses)
		result, err := client.CreateScanReport(scanID, ScanReportTypeXML, retry)

		expected := reportXML
		assert.NoError(t, err)
		assert.Equal(t, expected, string(result))
	})
	t.Run("fails if create report fails", func(t *testing.T) {
		scanID := 1000001
		postReportURL := "/CxRestAPI/help/reports/sastScan"
		getReportStatusURL := "/CxRestAPI/help/reports/sastScan/1251/status"
		getReportURL := "/CxRestAPI/help/reports/sastScan/1251"
		responses := map[string][]DoResponse{
			postReportURL: {
				{Response: makeResponse(500, "Internal error", "")}, //nolint:bodyclose
			},
			getReportStatusURL: {},
			getReportURL:       {},
		}
		client := makeCreateReportClient(responses)
		result, err := client.CreateScanReport(scanID, ScanReportTypeXML, retry)

		expectedErr := fmt.Sprintf("request POST %s/CxRestAPI/help/reports/sastScan failed with status code 500", BaseURL)
		assert.EqualError(t, err, expectedErr)
		assert.Equal(t, "", string(result))
	})
	t.Run("fails if get report status fails", func(t *testing.T) {
		scanID := 1000002
		postReportURL := "/CxRestAPI/help/reports/sastScan"
		getReportStatusURL := "/CxRestAPI/help/reports/sastScan/1252/status"
		getReportURL := "/CxRestAPI/help/reports/sastScan/1252"
		reportCreateJSON := `{
					"reportId": 1252, 
					"links": {
						"report": {"rel": "content", "uri": "/reports/sastScan/1252"},
						"status": {"rel": "status","uri": "/reports/sastScan/1252/status"}
					}
				}`
		responses := map[string][]DoResponse{
			postReportURL: {
				{Response: makeResponse(202, "Created", reportCreateJSON)}, //nolint:bodyclose
			},
			getReportStatusURL: {
				{Response: makeResponse(500, "Internal error", "")}, //nolint:bodyclose
			},
			getReportURL: {},
		}
		client := makeCreateReportClient(responses)
		result, err := client.CreateScanReport(scanID, ScanReportTypeXML, retry)

		expectedErr := fmt.Sprintf("request GET %s/CxRestAPI/help/reports/sastScan/1252/status failed with status code 500", BaseURL)
		assert.EqualError(t, err, expectedErr)
		assert.Equal(t, "", string(result))
	})
	t.Run("fails if get report status exhausts attempts ", func(t *testing.T) {
		scanID := 1000004
		postReportURL := "/CxRestAPI/help/reports/sastScan"
		getReportStatusURL := "/CxRestAPI/help/reports/sastScan/1254/status"
		getReportURL := "/CxRestAPI/help/reports/sastScan/1254"
		reportCreateJSON := `{
					"reportId": 1254, 
					"links": {
						"report": {"rel": "content", "uri": "/reports/sastScan/1254"},
						"status": {"rel": "status","uri": "/reports/sastScan/1254/status"}
					}
				}`
		reportStatusJSON := `{
					"link": {"rel": "content","uri": "/reports/sastScan/1254"},
					"contentType": "application/xml",
					"status": {"id": 2,"value": "In progress"}
				}`
		responses := map[string][]DoResponse{
			postReportURL: {
				{Response: makeResponse(202, "Created", reportCreateJSON)}, //nolint:bodyclose
			},
			getReportStatusURL: {
				{Response: makeOkResponse(reportStatusJSON)}, //nolint:bodyclose
				{Response: makeOkResponse(reportStatusJSON)}, //nolint:bodyclose
				{Response: makeOkResponse(reportStatusJSON)}, //nolint:bodyclose
			},
			getReportURL: {},
		}
		client := makeCreateReportClient(responses)
		result, err := client.CreateScanReport(scanID, ScanReportTypeXML, retry)

		assert.EqualError(t, err, "failed getting report after 3 attempts")
		assert.Equal(t, "", string(result))
	})
	t.Run("fails if fetch report fails", func(t *testing.T) {
		scanID := 1000003
		postReportURL := "/CxRestAPI/help/reports/sastScan"
		getReportStatusURL := "/CxRestAPI/help/reports/sastScan/1253/status"
		getReportURL := "/CxRestAPI/help/reports/sastScan/1253"
		reportCreateJSON := `{
					"reportId": 1253, 
					"links": {
						"report": {"rel": "content", "uri": "/reports/sastScan/1253"},
						"status": {"rel": "status","uri": "/reports/sastScan/1253/status"}
					}
				}`
		reportStatusJSON := `{
					"link": {"rel": "content","uri": "/reports/sastScan/1250"},
					"contentType": "application/xml",
					"status": {"id": 2,"value": "Created"}
				}`
		responses := map[string][]DoResponse{
			postReportURL: {
				{Response: makeResponse(202, "Created", reportCreateJSON)}, //nolint:bodyclose
			},
			getReportStatusURL: {
				{Response: makeOkResponse(reportStatusJSON)}, //nolint:bodyclose
			},
			getReportURL: {
				{Response: makeResponse(500, "Internal error", "")}, //nolint:bodyclose
			},
		}
		client := makeCreateReportClient(responses)
		result, err := client.CreateScanReport(scanID, ScanReportTypeXML, retry)

		expectedErr := fmt.Sprintf("request GET %s/CxRestAPI/help/reports/sastScan/1253 failed with status code 500", BaseURL)
		assert.EqualError(t, err, expectedErr)
		assert.Equal(t, "", string(result))
	})
}

func makeOkResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: 200,
		Status:     "OK",
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

func makeBadRequestResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: 400,
		Status:     "Bad Request",
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

func makeResponse(statusCode int, status, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     status,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

func makeCreateReportClient(responses map[string][]DoResponse) *APIClient {
	requestCounter := make(map[string]int, len(responses))
	for k := range responses {
		requestCounter[k] = 0
	}
	adapter := &HTTPClientMock2{DoHandler: func(request *retryablehttp.Request) (*http.Response, error) {
		url := request.URL.String()
		for k, i := range requestCounter {
			if strings.HasSuffix(url, k) {
				response := responses[k][i]
				requestCounter[k]++
				return response.Response, response.Err
			}
		}
		return nil, fmt.Errorf("unknown url %s", url)
	}}
	client, _ := NewSASTClient(BaseURL, adapter)
	client.Token = &AccessToken{AccessToken: "jwt", TokenType: "Bearer", ExpiresIn: 1234}
	return client
}
