//go:build integration

// Integration tests for issue #56 — Search.StartReindex (POST /search/reindex).
// Requires Blnk Core running at http://localhost:5001 with the master key.
//
// Run: go test -tags=integration -v ./integration/... -run Issue56
package integration

import (
	"net/http"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func intPtr(v int) *int {
	return &v
}

func TestIssue56_StartReindex(t *testing.T) {
	client := newIntegrationClient(t)

	started, resp, err := client.Search.StartReindex(nil)
	if err != nil && resp != nil && resp.StatusCode == http.StatusConflict {
		t.Skip("reindex already in progress on local Core")
	}
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusAccepted, resp.StatusCode)
	require.Equal(t, "Reindex operation started", started.Message)
	require.NotEmpty(t, started.Progress.Status)
}

func TestIssue56_StartReindex_WithBatchSize(t *testing.T) {
	client := newIntegrationClient(t)

	started, resp, err := client.Search.StartReindex(&blnkgo.StartReindexRequest{BatchSize: intPtr(500)})
	if err != nil && resp != nil && resp.StatusCode == http.StatusConflict {
		t.Skip("reindex already in progress on local Core")
	}
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusAccepted, resp.StatusCode)
	require.Equal(t, "Reindex operation started", started.Message)
}

func TestIssue56_StartReindex_ValidationError(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Search.StartReindex(&blnkgo.StartReindexRequest{BatchSize: intPtr(0)})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "batch_size must be a positive integer")
}
