//go:build integration

// Integration tests for LedgerBalance.Get with_queued query param.
// Requires Blnk Core running at http://localhost:5001 and BLNK_API_KEY in the environment.
//
// Run:
//
//	export BLNK_API_KEY=your_key
//	go test -tags=integration -v ./integration/... -run Issue127
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue127_GetBalance_WithQueued(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	ledger, resp, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{Name: "WithQueued " + suffix})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	bal, resp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID: ledger.LedgerID,
		Currency: "USD",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	withQueued, resp, err := client.LedgerBalance.Get(bal.BalanceID, &blnkgo.GetBalanceRequest{WithQueued: true})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, bal.BalanceID, withQueued.BalanceID)
}

func TestIssue127_GetBalance_WithFromSourceAndWithQueued(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	ledger, resp, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{Name: "WithQueuedCombo " + suffix})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	bal, resp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID: ledger.LedgerID,
		Currency: "USD",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	got, resp, err := client.LedgerBalance.Get(bal.BalanceID, &blnkgo.GetBalanceRequest{
		FromSource: true,
		WithQueued: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, bal.BalanceID, got.BalanceID)
}
