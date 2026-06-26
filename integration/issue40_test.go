//go:build integration

// Integration tests for issue #40 — Transaction.RecoverQueue (POST /transactions/recover).
// Requires Blnk Core running at http://localhost:5001 (docker compose up in blnk/).
//
// Run: go test -tags=integration -v ./integration/... -run Issue40
package integration

import (
	"net/http"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue40_RecoverQueue_Default(t *testing.T) {
	client := newIntegrationClient(t)

	result, resp, err := client.Transaction.RecoverQueue(blnkgo.RecoverQueueRequest{})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.GreaterOrEqual(t, result.Recovered, 0)
	require.NotEmpty(t, result.Threshold)
}

func TestIssue40_RecoverQueue_WithThreshold(t *testing.T) {
	client := newIntegrationClient(t)

	result, resp, err := client.Transaction.RecoverQueue(blnkgo.RecoverQueueRequest{
		Threshold: "24h",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.GreaterOrEqual(t, result.Recovered, 0)
	require.Equal(t, "24h0m0s", result.Threshold)
}

func TestIssue40_RecoverQueue_InvalidThreshold(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Transaction.RecoverQueue(blnkgo.RecoverQueueRequest{
		Threshold: "bogus",
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "threshold must be a valid duration string")
}
