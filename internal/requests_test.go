package internal

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

const (
	tokenURL = "/CxRestAPI/auth/identity/connect/token"
	username = "abcd"
	password = "Cx1234"
)

func TestRequests_CreateAccessTokenRequest(t *testing.T) {
	t.Run("create token successfully", func(t *testing.T) {
		request, err := CreateAccessTokenRequest(BaseURL, username, password)
		assert.NoError(t, err)
		assert.NotNil(t, request)
		assert.Equal(t, request.Method, http.MethodPost)
		assert.Equal(t, request.URL.Path, tokenURL)
	})
}

func TestRequests_CreateRequest(t *testing.T) {

	t.Run("create request successfully", func(t *testing.T) {
		response := http.Response{
			StatusCode: 200,
			Status:     "Success",
			Body:       nil,
		}
		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
		client, _ := NewSASTClient(BaseURL, adapter)

		err := client.Authenticate(username, password)

		request, err := CreateRequest(http.MethodGet, BaseURL, nil, client.Token)
		assert.NoError(t, err)
		assert.NotNil(t, request)
		assert.Equal(t, request.Method, http.MethodGet)
	})

	t.Run("create request unsuccessfully", func(t *testing.T) {
		response := http.Response{
			StatusCode: 400,
			Status:     "Bad Request",
			Body:       ioutil.NopCloser(bytes.NewBufferString(ErrorResponseJSON)),
		}
		adapter := &HTTPClientMock{DoResponse: response, DoError: nil}
		client, _ := NewSASTClient(BaseURL, adapter)

		errAuth := client.Authenticate(username, password)
		assert.Error(t, errAuth)

		request, errCreate := CreateRequest(http.MethodGet, BaseURL, nil, client.Token)
		assert.NoError(t, errCreate)
		assert.NotNil(t, request)
		assert.Equal(t, request.Method, http.MethodGet)
	})
}
