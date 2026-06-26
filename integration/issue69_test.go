//go:build integration

// Integration tests for issue #69 — LedgerBalance.Get from_source query param.
// Requires Blnk Core running at http://localhost:5001 and BLNK_API_KEY in the environment.
//
// Run:
//
//	export BLNK_API_KEY=your_key
//	go test -tags=integration -v ./integration/... -run Issue69
package integration

import (
	"fmt"
	"math/big"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue69_GetBalance_FromSourceMatchesSnapshot(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	ledger, resp, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{Name: "FromSource " + suffix})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	source, resp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID: ledger.LedgerID,
		Currency: "USD",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	dest, resp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID: ledger.LedgerID,
		Currency: "USD",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	_, resp, err = client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      500,
			Reference:   "from-source-" + suffix,
			Precision:   100,
			Currency:    "USD",
			Source:      source.BalanceID,
			Destination: dest.BalanceID,
			SkipQueue:   true,
			Description: "Fund destination for from_source test",
		},
		AllowOverdraft: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	snapshot, resp, err := client.LedgerBalance.Get(dest.BalanceID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	fromSource, resp, err := client.LedgerBalance.Get(dest.BalanceID, &blnkgo.GetBalanceRequest{FromSource: true})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	require.Equal(t, dest.BalanceID, fromSource.BalanceID)
	require.Equal(t, 0, snapshot.Balance.Cmp(fromSource.Balance))
	require.Equal(t, 0, snapshot.CreditBalance.Cmp(fromSource.CreditBalance))
	require.Equal(t, 0, snapshot.DebitBalance.Cmp(fromSource.DebitBalance))
	require.True(t, fromSource.Balance.Cmp(big.NewInt(0)) > 0, "destination should be credited")
}

func TestIssue69_GetBalance_BackwardCompatibleNoOptions(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	ledger, resp, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{Name: "GetCompat " + suffix})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	bal, resp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID: ledger.LedgerID,
		Currency: "USD",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	got, resp, err := client.LedgerBalance.Get(bal.BalanceID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, bal.BalanceID, got.BalanceID)
}
