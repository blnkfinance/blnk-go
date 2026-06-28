//go:build integration

// Integration tests for issue #54 — Hooks.Delete (DELETE /hooks/{id}).
// Requires Blnk Core running at http://localhost:5001 with the master key.
//
// Run: go test -tags=integration -v ./integration/... -run Issue54
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue54_DeleteHook(t *testing.T) {
	client := newIntegrationClient(t)

	created, createResp, err := client.Hooks.Create(blnkgo.CreateHookRequest{
		Name:       fmt.Sprintf("issue54-hook-%d", time.Now().UnixNano()),
		URL:        "https://api.example.com/validate",
		Type:       blnkgo.HookTypePreTransaction,
		Active:     true,
		Timeout:    30,
		RetryCount: 3,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	require.NotEmpty(t, created.ID)

	deleted, deleteResp, err := client.Hooks.Delete(created.ID)
	require.NoError(t, err)
	require.NotNil(t, deleteResp)
	require.Equal(t, http.StatusOK, deleteResp.StatusCode)
	require.Equal(t, "hook deleted successfully", deleted.Message)

	_, getResp, err := client.Hooks.Get(created.ID)
	require.Error(t, err)
	require.NotNil(t, getResp)
	require.NotEqual(t, http.StatusOK, getResp.StatusCode)
}

func TestIssue54_DeleteHook_ValidationError(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Hooks.Delete("")
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "hook id is required")
}

func TestIssue54_DeleteHook_NotFound(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Hooks.Delete("hook_nonexistent_issue54")
	require.Error(t, err)
	require.NotNil(t, resp)
	require.NotEqual(t, http.StatusOK, resp.StatusCode)
}
