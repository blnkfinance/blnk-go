//go:build integration

// Integration tests for issue #57 — Search.GetReindexStatus (GET /search/reindex).
// Requires Blnk Core running at http://localhost:5001 with the master key.
//
// Run: go test -tags=integration -v ./integration/... -run Issue57
package integration

import (
	"net/http"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue57_GetReindexStatus(t *testing.T) {
	client := newIntegrationClient(t)

	// Ensure a reindex has been started (may already be running from a prior test).
	_, startResp, err := client.Search.StartReindex(nil)
	if err != nil && startResp != nil && startResp.StatusCode == http.StatusConflict {
		// already in progress — fine for status check
	} else {
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, startResp.StatusCode)
	}

	progress, resp, err := client.Search.GetReindexStatus()
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotEmpty(t, progress.Status)
}

func TestIssue57_GetReindexStatus_AfterStart(t *testing.T) {
	client := newIntegrationClient(t)

	_, _, _ = client.Search.StartReindex(&blnkgo.StartReindexRequest{BatchSize: intPtr(1000)})

	progress, resp, err := client.Search.GetReindexStatus()
	if err != nil && resp != nil && resp.StatusCode == http.StatusNotFound {
		t.Skip("no reindex operation started on local Core")
	}
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotEmpty(t, progress.Status)
}
