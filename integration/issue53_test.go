//go:build integration

// Integration tests for issue #53 — Hooks.List (GET /hooks).
// Requires Blnk Core running at http://localhost:5001 with the master key.
//
// Run: go test -tags=integration -v ./integration/... -run Issue53
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue53_ListHooks(t *testing.T) {
	client := newIntegrationClient(t)

	created, createResp, err := client.Hooks.Create(blnkgo.CreateHookRequest{
		Name:       fmt.Sprintf("issue53-hook-%d", time.Now().UnixNano()),
		URL:        "https://api.example.com/validate",
		Type:       blnkgo.HookTypePreTransaction,
		Active:     true,
		Timeout:    30,
		RetryCount: 3,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	require.NotEmpty(t, created.ID)

	hooks, listResp, err := client.Hooks.List(&blnkgo.ListHooksOptions{Type: blnkgo.HookTypePreTransaction})
	require.NoError(t, err)
	require.NotNil(t, listResp)
	require.Equal(t, http.StatusOK, listResp.StatusCode)

	found := false
	for _, hook := range hooks {
		if hook.ID == created.ID {
			found = true
			require.Equal(t, created.Name, hook.Name)
			require.Equal(t, blnkgo.HookTypePreTransaction, hook.Type)
			break
		}
	}
	require.True(t, found, "expected created hook in list response")
}

func TestIssue53_ListHooks_WithoutOptions(t *testing.T) {
	client := newIntegrationClient(t)

	hooks, resp, err := client.Hooks.List(nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, hooks)
}

func TestIssue53_ListHooks_ValidationError(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Hooks.List(&blnkgo.ListHooksOptions{Type: "INVALID"})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "type must be PRE_TRANSACTION or POST_TRANSACTION")
}
