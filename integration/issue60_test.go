//go:build integration

// Integration tests for issue #60 — ApiKeys.Delete (DELETE /api-keys/{id}).
// Requires Blnk Core running at http://localhost:5001 with a master or api-keys:write key.
//
// Run: go test -tags=integration -v ./integration/... -run Issue60
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue60_DeleteApiKey(t *testing.T) {
	client := newIntegrationClient(t)

	owner := fmt.Sprintf("owner_issue60_%d", time.Now().UnixNano())
	keyName := fmt.Sprintf("issue60-key-%d", time.Now().UnixNano())
	expiresAt := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)

	created, createResp, err := client.ApiKeys.Create(blnkgo.CreateApiKeyRequest{
		Name:      keyName,
		Owner:     owner,
		Scopes:    []string{"ledgers:read"},
		ExpiresAt: expiresAt,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	require.NotEmpty(t, created.ApiKeyID)

	deleteResp, err := client.ApiKeys.Delete(created.ApiKeyID, &blnkgo.DeleteApiKeysOptions{Owner: owner})
	require.NoError(t, err)
	require.NotNil(t, deleteResp)
	require.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	keys, listResp, err := client.ApiKeys.List(&blnkgo.ListApiKeysOptions{Owner: owner})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, listResp.StatusCode)

	found := false
	for _, key := range keys {
		if key.ApiKeyID == created.ApiKeyID {
			found = true
			require.True(t, key.IsRevoked)
			break
		}
	}
	require.True(t, found, "expected revoked API key in list response")
}

func TestIssue60_DeleteApiKey_ValidationError(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.ApiKeys.Delete("", nil)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "api key id is required")
}

func TestIssue60_DeleteApiKey_NotFound(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.ApiKeys.Delete("api_key_nonexistent_issue60", &blnkgo.DeleteApiKeysOptions{Owner: "owner_issue60"})
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
