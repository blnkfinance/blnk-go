//go:build integration

// Integration tests for issue #123 — Transaction.List (GET /transactions).
// Requires Blnk Core running at http://localhost:5001 and BLNK_API_KEY in the environment.
//
// Run:
//
//	export BLNK_API_KEY=your_key
//	go test -tags=integration -v ./integration/... -run Issue123
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue123_ListTransactions(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	ref := "issue123-" + suffix

	created, resp, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      100,
			Reference:   ref,
			Precision:   100,
			Currency:    "USD",
			Source:      "@FundingPool",
			Destination: "@Issue123Dest",
			Description: "Issue123 list transaction",
			SkipQueue:   true,
		},
		AllowOverdraft: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, created.TransactionID)

	transactions, resp, err := client.Transaction.List()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, transactions)

	found := false
	for _, txn := range transactions {
		if txn.TransactionID == created.TransactionID {
			found = true
			require.Equal(t, ref, txn.Reference)
			break
		}
	}
	// Core defaults GET /transactions to limit=20. On a busy instance the new
	// transaction may fall outside that window; still require it when not full.
	if len(transactions) < 20 {
		require.True(t, found, "created transaction should appear in default list window")
	}
}
