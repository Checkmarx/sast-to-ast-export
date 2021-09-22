package internal

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	BaseURL                 = "http://127.0.0.1"
	ErrorResponseJSON       = `"error"`
	InvalidDataResponseJSON = `invalid data`
)

type HTTPClientMock struct {
	DoResponse http.Response
	DoError    error
}

func (c *HTTPClientMock) Do(req *http.Request) (*http.Response, error) {
	return &c.DoResponse, c.DoError
}

func TestNewSASTClient(t *testing.T) {
	response := http.Response{
		Body: ioutil.NopCloser(bytes.NewBufferString("test")),
	}
	adapter := &HTTPClientMock{DoResponse: response, DoError: nil}

	client, err := NewSASTClient(BaseURL, adapter)

	assert.NoError(t, err)
	assert.Equal(t, BaseURL, client.BaseURL)
	assert.Equal(t, adapter, client.Adapter)
}

func TestSASTClient_Authenticate(t *testing.T) {
	t.Run("authenticates successfully", func(t *testing.T) {
		responseJSON := `{"access_token":"jwt", "token_type":"Bearer", "expires_in": 1234}`
		response := makeOkResponse(responseJSON)
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
		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
		client, _ := NewSASTClient(BaseURL, adapter)

		err := client.Authenticate("username", "password")

		assert.Error(t, err)
		assert.Nil(t, client.Token)
	})

	t.Run("returns error if can't parse response", func(t *testing.T) {
		response := makeOkResponse(InvalidDataResponseJSON)
		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
		client, _ := NewSASTClient(BaseURL, adapter)

		err := client.Authenticate("username", "password")

		assert.Error(t, err)
		assert.Equal(t, &AccessToken{}, client.Token)
	})
}

func TestSASTClient_GetUsersResponseBody(t *testing.T) {
	mockToken := &AccessToken{AccessToken: "jwt", TokenType: "Bearer", ExpiresIn: 1234}

	t.Run("returns users response", func(t *testing.T) {
		responseJSON := `[{"id": 1, "userName": "test1", "lastLoginDate": "2021-08-17T12:22:28.2331383Z", "active": true},
						  {"id": 2, "userName": "test2", "lastLoginDate": "2021-08-17T12:22:28.2331383Z", "active": true},
						  {"id": 3, "userName": "test3", "lastLoginDate": "2021-08-17T12:22:28.2331383Z", "active": true}]`
		response := makeOkResponse(responseJSON)
		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
		client, _ := NewSASTClient(BaseURL, adapter)
		client.Token = mockToken

		result, err := client.GetUsers()

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, responseJSON, string(result))
	})

	t.Run("returns error if response is not HTTP OK", func(t *testing.T) {
		response := makeBadRequestResponse(ErrorResponseJSON)
		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
		client, _ := NewSASTClient(BaseURL, adapter)
		client.Token = mockToken

		result, err := client.GetUsers()

		assert.Error(t, err)
		assert.Len(t, result, 0)
	})
}

func TestSASTClient_GetTeamsResponseBody(t *testing.T) {
	mockToken := &AccessToken{AccessToken: "jwt", TokenType: "Bearer", ExpiresIn: 1234}

	t.Run("returns teams response", func(t *testing.T) {
		responseJSON := `[{"id": 1, "name": "test1", "fullName": "/CxServer/test1", "parentId": 1},
						  {"id": 2, "name": "test2", "fullName": "/CxServer/test2", "parentId": 1},
						  {"id": 3, "name": "test3", "fullName": "/CxServer/test3", "parentId": 1}]`
		response := makeOkResponse(responseJSON)
		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
		client, _ := NewSASTClient(BaseURL, adapter)
		client.Token = mockToken

		result, err := client.GetTeams()

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, responseJSON, string(result))
	})

	t.Run("returns error if response is not HTTP OK", func(t *testing.T) {
		response := makeBadRequestResponse(ErrorResponseJSON)
		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
		client, _ := NewSASTClient(BaseURL, adapter)
		client.Token = mockToken

		result, err := client.GetTeams()

		assert.Error(t, err)
		assert.Len(t, result, 0)
	})
}

func makeOkResponse(body string) http.Response {
	return http.Response{
		StatusCode: 200,
		Status:     "OK",
		Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
	}
}

func makeBadRequestResponse(body string) http.Response {
	return http.Response{
		StatusCode: 400,
		Status:     "Bad Request",
		Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
	}
}
