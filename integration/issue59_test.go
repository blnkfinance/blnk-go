//go:build integration

// Integration tests for issue #59 — ApiKeys.List (GET /api-keys).
// Requires Blnk Core running at http://localhost:5001 with a master or api-keys:read key.
//
// Run: go test -tags=integration -v ./integration/... -run Issue59
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue59_ListApiKeys(t *testing.T) {
	client := newIntegrationClient(t)

	owner := fmt.Sprintf("owner_issue59_%d", time.Now().UnixNano())
	keyName := fmt.Sprintf("issue59-key-%d", time.Now().UnixNano())
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

	keys, listResp, err := client.ApiKeys.List(&blnkgo.ListApiKeysOptions{Owner: owner})
	require.NoError(t, err)
	require.NotNil(t, listResp)
	require.Equal(t, http.StatusOK, listResp.StatusCode)
	require.NotEmpty(t, keys)

	found := false
	for _, key := range keys {
		if key.ApiKeyID == created.ApiKeyID {
			found = true
			require.Equal(t, owner, key.OwnerID)
			require.Equal(t, keyName, key.Name)
			require.Equal(t, []string{"ledgers:read"}, key.Scopes)
			require.False(t, key.IsRevoked)
			break
		}
	}
	require.True(t, found, "expected created API key in list response")
}

func TestIssue59_ListApiKeys_ValidationError(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.ApiKeys.List(&blnkgo.ListApiKeysOptions{Owner: "   "})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "owner must be a non-empty string")
}
