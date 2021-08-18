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
		responseJSON := `{"access_token":"jwttoken", "token_type":"Bearer", "expires_in": 1234}`
		response := http.Response{
			StatusCode: 200,
			Status:     "OK",
			Body:       ioutil.NopCloser(bytes.NewBufferString(responseJSON)),
		}
		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
		client, _ := NewSASTClient(BaseURL, adapter)

		err := client.Authenticate("username", "password")

		assert.NoError(t, err)
		assert.NotNil(t, client.Token)
		expected := &AccessToken{AccessToken: "jwttoken", TokenType: "Bearer", ExpiresIn: 1234}
		assert.Equal(t, client.Token, expected)
	})

	t.Run("returns error if response is not HTTP OK", func(t *testing.T) {
		response := http.Response{
			StatusCode: 400,
			Status:     "Bad Request",
			Body:       ioutil.NopCloser(bytes.NewBufferString(ErrorResponseJSON)),
		}
		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
		client, _ := NewSASTClient(BaseURL, adapter)

		err := client.Authenticate("username", "password")

		assert.Error(t, err)
		assert.Nil(t, client.Token)
	})

	t.Run("returns error if can't parse response", func(t *testing.T) {
		response := http.Response{
			StatusCode: 200,
			Status:     "OK",
			Body:       ioutil.NopCloser(bytes.NewBufferString(InvalidDataResponseJSON)),
		}
		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
		client, _ := NewSASTClient(BaseURL, adapter)

		err := client.Authenticate("username", "password")

		assert.Error(t, err)
		assert.Equal(t, &AccessToken{}, client.Token)
	})
}

// outdated test
// kept for reference
//func TestSASTClient_GetProjects(t *testing.T) {
//	mockToken := &AccessToken{AccessToken: "jwt", TokenType: "Bearer", ExpiresIn: 1234}
//
//	t.Run("returns project list", func(t *testing.T) {
//		responseJSON := `[{"id": 1, "teamId": 1, "name": "project1", "isPublic": true},
//						  {"id": 2, "teamId": 1, "name": "project2", "isPublic": false},
//						  {"id": 3, "teamId": 2, "name": "project3", "isPublic": true}]`
//		response := http.Response{
//			StatusCode: 200,
//			Status:     "OK",
//			Body:       ioutil.NopCloser(bytes.NewBufferString(responseJSON)),
//		}
//		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
//		client, _ := NewSASTClient(BaseURL, adapter)
//		client.Token = mockToken
//
//		result, err := client.GetProjects()
//
//		assert.NoError(t, err)
//		assert.NotNil(t, result)
//		expected := []Project{
//			{ID: 1, TeamID: 1, Name: "project1", IsPublic: true},
//			{ID: 2, TeamID: 1, Name: "project2", IsPublic: false},
//			{ID: 3, TeamID: 2, Name: "project3", IsPublic: true},
//		}
//		assert.Equal(t, expected, result)
//	})
//
//	t.Run("returns error if response is not HTTP OK", func(t *testing.T) {
//		response := http.Response{
//			StatusCode: 400,
//			Status:     "Bad Request",
//			Body:       ioutil.NopCloser(bytes.NewBufferString(ErrorResponseJSON)),
//		}
//		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
//		client, _ := NewSASTClient(BaseURL, adapter)
//		client.Token = mockToken
//
//		result, err := client.GetProjects()
//
//		assert.Error(t, err)
//		assert.Len(t, result, 0)
//	})
//
//	t.Run("returns error if can't parse response", func(t *testing.T) {
//		response := http.Response{
//			StatusCode: 200,
//			Status:     "OK",
//			Body:       ioutil.NopCloser(bytes.NewBufferString(InvalidDataResponseJSON)),
//		}
//		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
//		client, _ := NewSASTClient(BaseURL, adapter)
//		client.Token = mockToken
//
//		result, err := client.GetProjects()
//
//		assert.Error(t, err)
//		assert.Len(t, result, 0)
//	})
//}
