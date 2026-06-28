//go:build integration

// Integration tests for issue #61 — Health.Check (GET /health).
// Requires Blnk Core running at http://localhost:5001 (docker compose up in blnk/).
//
// Run: go test -tags=integration -v ./integration/... -run Issue61
package integration

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func newHealthClient(t *testing.T) *blnkgo.Client {
	t.Helper()
	u := integrationBaseURL(t)
	// Health does not require authentication; API key is optional.
	return blnkgo.NewClient(u, nil, blnkgo.WithTimeout(15*time.Second), blnkgo.WithRetry(2))
}

func TestIssue61_HealthCheck(t *testing.T) {
	client := newHealthClient(t)

	health, resp, err := client.Health.Check()
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, health)
	require.Equal(t, "UP", health.Status)
}

func TestIssue61_HealthCheck_WithAPIKey(t *testing.T) {
	client := newIntegrationClient(t)

	health, resp, err := client.Health.Check()
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "UP", health.Status)
}

func TestIssue61_HealthCheck_InvalidBaseURL(t *testing.T) {
	u, err := url.Parse("http://localhost:59999/")
	require.NoError(t, err)

	client := blnkgo.NewClient(u, nil, blnkgo.WithTimeout(2*time.Second), blnkgo.WithRetry(1))
	_, resp, err := client.Health.Check()
	require.Error(t, err)
	if resp != nil {
		require.NotEqual(t, http.StatusOK, resp.StatusCode)
	}
}
