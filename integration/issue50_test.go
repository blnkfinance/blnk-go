//go:build integration

// Integration tests for issue #50 — Hooks.Create (POST /hooks).
// Requires Blnk Core running at http://localhost:5001 with the master key.
//
// Run: go test -tags=integration -v ./integration/... -run Issue50
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue50_CreateHook(t *testing.T) {
	client := newIntegrationClient(t)

	hook, resp, err := client.Hooks.Create(blnkgo.CreateHookRequest{
		Name:       fmt.Sprintf("issue50-hook-%d", time.Now().UnixNano()),
		URL:        "https://api.example.com/validate",
		Type:       blnkgo.HookTypePreTransaction,
		Active:     true,
		Timeout:    30,
		RetryCount: 3,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotNil(t, hook)
	require.NotEmpty(t, hook.ID)
	require.Equal(t, blnkgo.HookTypePreTransaction, hook.Type)
	require.Equal(t, "https://api.example.com/validate", hook.URL)
	require.True(t, hook.Active)
	require.Equal(t, 30, hook.Timeout)
	require.Equal(t, 3, hook.RetryCount)
	require.NotEmpty(t, hook.CreatedAt)
}

func TestIssue50_CreateHook_ValidationError(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Hooks.Create(blnkgo.CreateHookRequest{
		Name:       "invalid-hook",
		Type:       blnkgo.HookTypePreTransaction,
		Active:     true,
		Timeout:    30,
		RetryCount: 3,
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "url is required")
}
