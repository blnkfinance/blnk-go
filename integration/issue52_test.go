//go:build integration

// Integration tests for issue #52 — Hooks.Get (GET /hooks/{id}).
// Requires Blnk Core running at http://localhost:5001 with the master key.
//
// Run: go test -tags=integration -v ./integration/... -run Issue52
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue52_GetHook(t *testing.T) {
	client := newIntegrationClient(t)

	created, createResp, err := client.Hooks.Create(blnkgo.CreateHookRequest{
		Name:       fmt.Sprintf("issue52-hook-%d", time.Now().UnixNano()),
		URL:        "https://api.example.com/validate",
		Type:       blnkgo.HookTypePreTransaction,
		Active:     true,
		Timeout:    30,
		RetryCount: 3,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	require.NotEmpty(t, created.ID)

	hook, getResp, err := client.Hooks.Get(created.ID)
	require.NoError(t, err)
	require.NotNil(t, getResp)
	require.Equal(t, http.StatusOK, getResp.StatusCode)
	require.Equal(t, created.ID, hook.ID)
	require.Equal(t, created.Name, hook.Name)
	require.Equal(t, created.URL, hook.URL)
	require.Equal(t, created.Type, hook.Type)
	require.Equal(t, created.Active, hook.Active)
	require.Equal(t, created.Timeout, hook.Timeout)
	require.Equal(t, created.RetryCount, hook.RetryCount)
}

func TestIssue52_GetHook_ValidationError(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Hooks.Get("")
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "hook id is required")
}

func TestIssue52_GetHook_NotFound(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Hooks.Get("hk_nonexistent_issue52")
	require.Error(t, err)
	require.NotNil(t, resp)
	require.NotEqual(t, http.StatusOK, resp.StatusCode)
}
