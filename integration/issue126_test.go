//go:build integration

// Integration tests for issue #126 — Core 0.15.0 dropped response fields.
// Requires Blnk Core 0.15.0+ at http://localhost:5001 and BLNK_API_KEY.
//
// Run:
//
//	export BLNK_API_KEY=your_key
//	go test -tags=integration -v ./integration/... -run Issue126
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue126_DroppedResponseFields(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	ledger, resp, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{
		Name: "Issue126 Dropped Fields " + suffix,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	balance, resp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID: ledger.LedgerID,
		Currency: "USD",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, 0.0, balance.CurrencyMultiplier, "Core 0.15.0+ omits currency_multiplier")

	gotBalance, resp, err := client.LedgerBalance.Get(balance.BalanceID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, 0.0, gotBalance.CurrencyMultiplier)

	txn, resp, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      25,
			Reference:   "issue126-" + suffix,
			Precision:   100,
			Currency:    "USD",
			Source:      "@FundingPool",
			Destination: "@Issue126Dest",
			Description: "Issue126 dropped fields probe",
			SkipQueue:   true,
			Rate:        1.1, // source-compatible create literal
		},
		AllowOverdraft: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, 0.0, txn.Rate, "Core 0.15.0+ omits rate from responses")
}
