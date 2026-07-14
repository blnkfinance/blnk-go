//go:build integration

// Integration tests for issue #125 — Search.MultiSearch (POST /multi-search).
// Requires Blnk Core running at http://localhost:5001 with Typesense available,
// and BLNK_API_KEY in the environment.
//
// Run:
//
//	export BLNK_API_KEY=your_key
//	go test -tags=integration -v ./integration/... -run Issue125
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue125_MultiSearch(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	ledgerName := "Issue125 MultiSearch " + suffix

	ledger, resp, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{Name: ledgerName})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, ledger.LedgerID)

	deadline := time.Now().Add(30 * time.Second)
	var multiResp *blnkgo.MultiSearchResponse
	for {
		multiResp, resp, err = client.Search.MultiSearch(blnkgo.MultiSearchRequest{
			Searches: []blnkgo.MultiSearchCollectionParams{
				{
					Collection: "ledgers",
					Q:          ledgerName,
					QueryBy:    "name",
					PerPage:    5,
				},
				{
					Collection: "balances",
					Q:          "*",
					QueryBy:    "currency",
					PerPage:    1,
				},
			},
		})
		if err == nil && resp.StatusCode == http.StatusOK && len(multiResp.Results) == 2 && multiResp.Results[0].Found > 0 {
			break
		}
		if time.Now().After(deadline) {
			require.NoError(t, err, "multi-search timed out waiting for index")
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.Len(t, multiResp.Results, 2)
			require.Greater(t, multiResp.Results[0].Found, 0, "expected indexed ledger in multi-search results")
		}
		time.Sleep(500 * time.Millisecond)
	}

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, multiResp.Results, 2)
	require.Greater(t, multiResp.Results[0].Found, 0)
	require.GreaterOrEqual(t, multiResp.Results[1].Found, 0)
}

func TestIssue125_MultiSearch_EmptySearches(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Search.MultiSearch(blnkgo.MultiSearchRequest{})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "searches cannot be empty")
}
