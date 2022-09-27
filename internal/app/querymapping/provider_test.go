package querymapping

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type HTTPClientMock struct {
	GetResponse *http.Response
	GetError    error
}

func (c *HTTPClientMock) Get(url string) (*http.Response, error) {
	return c.GetResponse, c.GetError
}

func TestQueryMappingProvider(t *testing.T) {
	t.Run("Test creating from file", func(t *testing.T) {
		response := http.Response{}
		adapter := &HTTPClientMock{GetResponse: &response, GetError: nil}

		provider, err := NewProvider("../../../data/mapping.json", adapter)
		assert.NoError(t, err)

		assert.Equal(t, "11", provider.GetMapping()[0].SastID)
	})

	t.Run("Test creating from URL", func(t *testing.T) {
		response := http.Response{
			Body: io.NopCloser(bytes.NewBufferString("{ \"mappings\": [{ \"astID\": \"5667386434418802377\",	\"sastID\": \"11\"}] }")),
		}
		adapter := &HTTPClientMock{GetResponse: &response, GetError: nil}

		provider, err := NewProvider("https://raw.githubusercontent.com/mapping.json", adapter)
		assert.NoError(t, err)

		assert.Equal(t, "11", provider.GetMapping()[0].SastID)
	})

	t.Run("Test error with wrong path", func(t *testing.T) {
		response := http.Response{}
		adapter := &HTTPClientMock{GetResponse: &response, GetError: nil}
		var err error
		_, err = NewProvider("wrong_path", adapter)
		assert.Error(t, err)
	})
}
