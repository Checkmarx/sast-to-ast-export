package internal

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestRequests_CreateAccessTokenRequest(t *testing.T) {
	tokenURL := "/CxRestAPI/auth/identity/connect/token"
	username := "abcd"
	password := "Cx1234"
	data := "client_id=resource_owner_client&client_secret=014DF517-39D1-4453-B7B3-9930C563627C&grant_type=password&password=%s&scope=sast_rest_api+access_control_api&username=%s"

	t.Run("create token successfully", func(t *testing.T) {
		body, _ := fmt.Printf(data, password, username)
		request, err := CreateAccessTokenRequest(BaseURL, username, password)
		assert.NoError(t, err)
		assert.NotNil(t, request)
		assert.Equal(t, request.Method, http.MethodPost)
		assert.Equal(t, request.URL.Path, tokenURL)
		assert.Equal(t, request.Body, body)
	})

	t.Run("create token not successfully", func(t *testing.T) {
		//body, _ := fmt.Printf(data, password, username)
		request, err := CreateAccessTokenRequest(BaseURL, "", "")
		assert.NoError(t, err)
		assert.NotNil(t, request)
		assert.Equal(t, request.Method, http.MethodPost)
		assert.Equal(t, request.URL.Path, tokenURL)
		//assert.Equal(t, request.Body, body)
	})
}
