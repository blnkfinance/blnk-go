//go:build integration

package integration

import (
	"net/url"
	"os"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func integrationAPIKey(t *testing.T) string {
	t.Helper()
	key := os.Getenv("BLNK_API_KEY")
	if key == "" {
		t.Skip("BLNK_API_KEY is not set; skipping integration test against Blnk Core")
	}
	return key
}

func integrationBaseURL(t *testing.T) *url.URL {
	t.Helper()
	raw := os.Getenv("BLNK_BASE_URL")
	if raw == "" {
		raw = "http://localhost:5001/"
	}
	u, err := url.Parse(raw)
	require.NoError(t, err)
	return u
}

func newIntegrationClient(t *testing.T) *blnkgo.Client {
	t.Helper()
	apiKey := integrationAPIKey(t)
	u := integrationBaseURL(t)
	return blnkgo.NewClient(u, &apiKey, blnkgo.WithTimeout(15*time.Second), blnkgo.WithRetry(2))
}
