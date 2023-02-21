package clients

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blockwarecom/insight-api/model"
	"github.com/stretchr/testify/assert"
)

func TestClient_GetRegistryCurrent(t *testing.T) {
	t.Run("check that the bearer token is sent in the request header, we can decode the model", func(t *testing.T) {
		// Create a mock HTTP server to simulate the registry API endpoint
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("Expected 'GET' request, got '%v'", r.Method)
			}
			if r.URL.Path != "/v1/registry/handle/name/current" {
				t.Errorf("Expected URL path '/v1/registry/handle/name/current', got '%v'", r.URL.Path)
			}
			if r.Header.Get("Authorization") != "Bearer test_token" {
				t.Errorf("Expected 'Authorization' header value 'Bearer test_token', got '%v'", r.Header.Get("Authorization"))
			}

			// Return a sample response
			version := &model.PublicAssetVersion{
				Version: "1.0.0",
				Content: model.AssetContent{
					Kind: "plan",
					Metadata: model.Metadata{
						Name:        "handle/name",
						Title:       "display_name",
						Description: "description",
					},
				},
			}
			_ = json.NewEncoder(w).Encode(version)
		}))
		defer server.Close()

		// Create a new client with the mock server URL and test token
		client := &Client{
			ServerURL: server.URL,
			Token:     "test_token",
		}

		// Call the method to get the current version
		version, err := client.GetRegistryCurrent("handle", "name")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Verify that the returned version matches the expected value
		assert.Equal(t, "1.0.0", version.Version)
		assert.Equal(t, "plan", version.Content.Kind)
		assert.Equal(t, "handle/name", version.Content.Metadata.Name)
		assert.Equal(t, "display_name", version.Content.Metadata.Title)
		assert.Equal(t, "description", version.Content.Metadata.Description)
	})
}
