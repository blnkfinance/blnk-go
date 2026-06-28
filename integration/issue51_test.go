//go:build integration

// Integration tests for issue #51 — Hooks.Update (PUT /hooks/{id}).
// Requires Blnk Core running at http://localhost:5001 with the master key.
//
// Run: go test -tags=integration -v ./integration/... -run Issue51
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue51_UpdateHook(t *testing.T) {
	client := newIntegrationClient(t)

	created, createResp, err := client.Hooks.Create(blnkgo.CreateHookRequest{
		Name:       fmt.Sprintf("issue51-hook-%d", time.Now().UnixNano()),
		URL:        "https://api.example.com/validate",
		Type:       blnkgo.HookTypePreTransaction,
		Active:     true,
		Timeout:    30,
		RetryCount: 3,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	require.NotEmpty(t, created.ID)

	updatedName := fmt.Sprintf("issue51-updated-%d", time.Now().UnixNano())
	updated, updateResp, err := client.Hooks.Update(created.ID, blnkgo.UpdateHookRequest{
		Name:       updatedName,
		URL:        "https://api.example.com/validate-v2",
		Type:       blnkgo.HookTypePreTransaction,
		Active:     false,
		Timeout:    45,
		RetryCount: 5,
	})
	require.NoError(t, err)
	require.NotNil(t, updateResp)
	require.Equal(t, http.StatusOK, updateResp.StatusCode)
	require.Equal(t, created.ID, updated.ID)
	require.Equal(t, updatedName, updated.Name)
	require.Equal(t, "https://api.example.com/validate-v2", updated.URL)
	require.False(t, updated.Active)
	require.Equal(t, 45, updated.Timeout)
	require.Equal(t, 5, updated.RetryCount)
}

func TestIssue51_UpdateHook_ValidationError(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Hooks.Update("", blnkgo.UpdateHookRequest{
		Name:       "updated",
		URL:        "https://api.example.com/validate",
		Type:       blnkgo.HookTypePreTransaction,
		Active:     true,
		Timeout:    30,
		RetryCount: 3,
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "hook id is required")
}

func TestIssue51_UpdateHook_NotFound(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Hooks.Update("hk_nonexistent_issue51", blnkgo.UpdateHookRequest{
		Name:       "updated",
		URL:        "https://api.example.com/validate",
		Type:       blnkgo.HookTypePreTransaction,
		Active:     true,
		Timeout:    30,
		RetryCount: 3,
	})
	require.Error(t, err)
	require.NotNil(t, resp)
	require.NotEqual(t, http.StatusOK, resp.StatusCode)
}
