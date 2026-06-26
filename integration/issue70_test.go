//go:build integration

// Integration tests for issue #70 — Transaction.Refund optional skip_queue body.
// Requires Blnk Core running at http://localhost:5001 (docker compose up in blnk/).
//
// Run: go test -tags=integration -v ./integration/... -run Issue70
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func createAppliedTransaction(t *testing.T, client *blnkgo.Client, refPrefix string) string {
	t.Helper()
	ref := fmt.Sprintf("%s-%d", refPrefix, time.Now().UnixNano())

	txn, resp, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      500,
			Reference:   ref,
			Precision:   100,
			Currency:    "USD",
			Source:      "@FundingPool",
			Destination: "@Recipient",
			Description: "Refund integration source tx",
			SkipQueue:   true,
		},
		AllowOverdraft: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, txn.TransactionID)
	require.Equal(t, blnkgo.PryTransactionStatus("APPLIED"), txn.Status)
	return txn.TransactionID
}

func TestIssue70_Refund_QueuedDefault(t *testing.T) {
	client := newIntegrationClient(t)
	originalTxnID := createAppliedTransaction(t, client, "issue70-queued")

	refund, resp, err := client.Transaction.Refund(originalTxnID, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, refund.TransactionID)
}

func TestIssue70_Refund_SkipQueueSynchronous(t *testing.T) {
	client := newIntegrationClient(t)
	originalTxnID := createAppliedTransaction(t, client, "issue70-sync")

	refund, resp, err := client.Transaction.Refund(originalTxnID, &blnkgo.RefundTransactionRequest{
		SkipQueue: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, refund.TransactionID)
	require.Equal(t, blnkgo.PryTransactionStatus("APPLIED"), refund.Status)
}
