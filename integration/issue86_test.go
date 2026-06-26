//go:build integration

// Integration tests for issue #86 — skip_queue on inflight commit/void (UpdateStatus + bulk).
// Requires Blnk Core running at http://localhost:5001 and BLNK_API_KEY in the environment.
//
// Run:
//
//	export BLNK_API_KEY=your_key
//	go test -tags=integration -v ./integration/... -run Issue86
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func createInflightTxnIssue86(t *testing.T, client *blnkgo.Client, ref string) string {
	t.Helper()
	expiry := time.Now().Add(24 * time.Hour)
	txn, resp, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      5000,
			Reference:   ref,
			Precision:   100,
			Currency:    "USD",
			Source:      "@FundingPool",
			Destination: "@Recipient",
			Description: "Issue 86 inflight setup",
			SkipQueue:   true,
		},
		Inflight:           true,
		InflightExpiryDate: &expiry,
		AllowOverdraft:     true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, txn.TransactionID)
	return txn.TransactionID
}

func TestIssue86_UpdateStatus_SkipQueueSynchronousCommit(t *testing.T) {
	client := newIntegrationClient(t)
	ref := fmt.Sprintf("issue86-sync-commit-%d", time.Now().UnixNano())
	txnID := createInflightTxnIssue86(t, client, ref)

	committed, resp, err := client.Transaction.Update(txnID, blnkgo.UpdateStatus{
		Status:    blnkgo.InflightStatusCommit,
		SkipQueue: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, blnkgo.PryTransactionStatus("APPLIED"), committed.Status)
	require.False(t, committed.Queued)
}

func TestIssue86_UpdateStatus_QueuedDefaultCommit(t *testing.T) {
	client := newIntegrationClient(t)
	ref := fmt.Sprintf("issue86-queued-commit-%d", time.Now().UnixNano())
	txnID := createInflightTxnIssue86(t, client, ref)

	committed, resp, err := client.Transaction.Update(txnID, blnkgo.UpdateStatus{
		Status: blnkgo.InflightStatusCommit,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Core 0.15+ queues commit by default (queued:true, status INFLIGHT).
	// Core 0.14.x may apply synchronously (APPLIED). Accept either contract.
	if committed.Queued {
		require.Equal(t, blnkgo.PryTransactionStatus("INFLIGHT"), committed.Status)
	} else {
		require.Equal(t, blnkgo.PryTransactionStatus("APPLIED"), committed.Status)
	}
}

func TestIssue86_UpdateStatus_SkipQueueSynchronousVoid(t *testing.T) {
	client := newIntegrationClient(t)
	ref := fmt.Sprintf("issue86-sync-void-%d", time.Now().UnixNano())
	txnID := createInflightTxnIssue86(t, client, ref)

	voided, resp, err := client.Transaction.Update(txnID, blnkgo.UpdateStatus{
		Status:    blnkgo.InflightStatusVoid,
		SkipQueue: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, blnkgo.PryTransactionStatus("VOID"), voided.Status)
}

func TestIssue86_BulkCommitInflight_SkipQueue(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	syncID := createInflightTxnIssue86(t, client, fmt.Sprintf("issue86-bulk-sync-%s", suffix))

	result, resp, err := client.Transaction.BulkCommitInflight(blnkgo.BulkCommitInflightRequest{
		SkipQueue: true,
		Transactions: []blnkgo.BulkCommitInflightItem{
			{TransactionID: syncID},
		},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, 1, result.Succeeded)
	require.Equal(t, "succeeded", result.Results[0].Status)
}

func TestIssue86_BulkVoidInflight_SkipQueue(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	syncID := createInflightTxnIssue86(t, client, fmt.Sprintf("issue86-bulk-void-%s", suffix))

	result, resp, err := client.Transaction.BulkVoidInflight(blnkgo.BulkVoidInflightRequest{
		SkipQueue:      true,
		TransactionIDs: []string{syncID},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, 1, result.Succeeded)
	require.Equal(t, "succeeded", result.Results[0].Status)
}
