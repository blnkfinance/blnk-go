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

	refund, resp, err := client.Transaction.Refund(originalTxnID)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, refund.TransactionID)
	require.Equal(t, originalTxnID, refund.ParentTransactionID)
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
	require.Equal(t, originalTxnID, refund.ParentTransactionID)
}

func TestIssue70_Refund_ReversesSourceAndDestination(t *testing.T) {
	client := newIntegrationClient(t)
	ref := fmt.Sprintf("issue70-reversal-%d", time.Now().UnixNano())

	original, resp, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      500,
			Reference:   ref,
			Precision:   100,
			Currency:    "USD",
			Source:      "@FundingPool",
			Destination: "@Recipient",
			Description: "Refund reversal integration tx",
			SkipQueue:   true,
		},
		AllowOverdraft: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	refund, resp, err := client.Transaction.Refund(original.TransactionID, &blnkgo.RefundTransactionRequest{
		SkipQueue: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, original.Amount, refund.Amount)
	require.Equal(t, original.Destination, refund.Source)
	require.Equal(t, original.Source, refund.Destination)
	require.Equal(t, original.TransactionID, refund.ParentTransactionID)
}

func TestIssue70_Refund_DuplicateRefundRejected(t *testing.T) {
	client := newIntegrationClient(t)
	originalTxnID := createAppliedTransaction(t, client, "issue70-dup")

	_, resp, err := client.Transaction.Refund(originalTxnID, &blnkgo.RefundTransactionRequest{
		SkipQueue: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	_, dupResp, dupErr := client.Transaction.Refund(originalTxnID, &blnkgo.RefundTransactionRequest{
		SkipQueue: true,
	})
	require.Error(t, dupErr)
	require.NotNil(t, dupResp)
	require.NotEqual(t, http.StatusCreated, dupResp.StatusCode)
}
