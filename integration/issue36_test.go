//go:build integration

// Integration tests for issue #36 — Transaction.GetLineage (GET /transactions/{id}/lineage).
// Requires Blnk Core running at http://localhost:5001 and BLNK_API_KEY in the environment.
// Works on Core 0.14.x+ (fund lineage); does not require 0.15-only features.
//
// Run:
//
//	export BLNK_API_KEY=your_key
//	go test -tags=integration -v ./integration/... -run Issue36
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func newClientIssue36(t *testing.T) *blnkgo.Client {
	return newIntegrationClient(t)
}

func TestIssue36_GetTransactionLineage(t *testing.T) {
	client := newClientIssue36(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Seed real balances; @FundingPool/@Recipient internal references are not
	// reliably resolvable across Core setups, so use concrete balance IDs.
	ledger, resp, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{Name: "Lineage Test " + suffix})
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

	// skip_queue applies synchronously so the transaction is immediately retrievable.
	txn, resp, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      100,
			Reference:   "lineage-" + suffix,
			Precision:   100,
			Currency:    "USD",
			Source:      source.BalanceID,
			Destination: dest.BalanceID,
			SkipQueue:   true,
			Description: "GetLineage integration tx",
		},
		AllowOverdraft: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, txn.TransactionID)

	lineage, resp, err := client.Transaction.GetLineage(txn.TransactionID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, txn.TransactionID, lineage.TransactionID)
}

func TestIssue36_GetTransactionLineage_EmptyID(t *testing.T) {
	client := newClientIssue36(t)

	lineage, resp, err := client.Transaction.GetLineage("")
	require.Error(t, err)
	require.Nil(t, lineage)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "transactionID is required")
}
