//go:build integration

// Integration tests for issue #122 — LedgerBalance.List (GET /balances).
// Requires Blnk Core running at http://localhost:5001 and BLNK_API_KEY in the environment.
//
// Run:
//
//	export BLNK_API_KEY=your_key
//	go test -tags=integration -v ./integration/... -run Issue122
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue122_ListBalances(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	ledger, resp, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{Name: "Issue122 List " + suffix})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	created, resp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID: ledger.LedgerID,
		Currency: "USD",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, created.BalanceID)

	balances, resp, err := client.LedgerBalance.List()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, balances)

	found := false
	for _, bal := range balances {
		if bal.BalanceID == created.BalanceID {
			found = true
			require.Equal(t, ledger.LedgerID, bal.LedgerID)
			require.Equal(t, "USD", bal.Currency)
			break
		}
	}
	// Core defaults GET /balances to limit=10. On a busy instance the new balance may
	// fall outside that window; still require it when the window is not full.
	if len(balances) < 10 {
		require.True(t, found, "created balance should appear in default list window")
	}
}
