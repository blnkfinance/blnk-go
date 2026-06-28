//go:build integration

// Integration tests for issue #58 — ApiKeys.Create (POST /api-keys).
// Requires Blnk Core running at http://localhost:5001 with a master or api-keys:write key.
//
// Run: go test -tags=integration -v ./integration/... -run Issue58
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue58_CreateApiKey(t *testing.T) {
	client := newIntegrationClient(t)

	owner := fmt.Sprintf("owner_issue58_%d", time.Now().UnixNano())
	expiresAt := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)

	apiKey, resp, err := client.ApiKeys.Create(blnkgo.CreateApiKeyRequest{
		Name:      fmt.Sprintf("issue58-key-%d", time.Now().UnixNano()),
		Owner:     owner,
		Scopes:    []string{"ledgers:read"},
		ExpiresAt: expiresAt,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotNil(t, apiKey)
	require.NotEmpty(t, apiKey.ApiKeyID)
	require.Contains(t, apiKey.ApiKeyID, "api_key_")
	require.NotEmpty(t, apiKey.Key)
	require.Equal(t, owner, apiKey.OwnerID)
	require.Equal(t, []string{"ledgers:read"}, apiKey.Scopes)
	require.False(t, apiKey.IsRevoked)
}

func TestIssue58_CreateApiKey_ValidationError(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.ApiKeys.Create(blnkgo.CreateApiKeyRequest{
		Name:  "invalid",
		Owner: "owner_test",
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "at least one scope must be specified")
}
