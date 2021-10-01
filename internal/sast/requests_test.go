package sast

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	tokenURL = "/CxRestAPI/auth/identity/connect/token" //nolint:gosec
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
		responseContent := "{\"access_token\": \"eyJhbGcOiJ...YRfzdQ\",\"expires_in\": 86400,\"token_type\": \"Bearer\"}"
		response := http.Response{
			StatusCode: 200,
			Status:     "Success",
			Body:       io.NopCloser(strings.NewReader(responseContent)),
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
