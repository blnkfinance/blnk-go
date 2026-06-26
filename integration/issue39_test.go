//go:build integration

// Integration tests for issue #39 — Transaction.BulkVoidInflight (POST /transactions/inflight/bulk/void).
// Requires Blnk Core running at http://localhost:5001 (docker compose up in blnk/).
//
// Run: go test -tags=integration -v ./integration/... -run Issue39
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func newClientIssue39(t *testing.T) *blnkgo.Client {
	return newIntegrationClient(t)
}

func createInflightTxnIssue39(t *testing.T, client *blnkgo.Client, ref string, amount float64) string {
	t.Helper()
	expiry := time.Now().Add(24 * time.Hour)
	txn, resp, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      amount,
			Reference:   ref,
			Precision:   100,
			Currency:    "USD",
			Source:      "@FundingPool",
			Destination: "@Recipient",
			Description: "Bulk void inflight setup",
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

func TestIssue39_BulkVoidInflight(t *testing.T) {
	client := newClientIssue39(t)
	refPrefix := fmt.Sprintf("bulk-void-%d", time.Now().UnixNano())

	id1 := createInflightTxnIssue39(t, client, refPrefix+"-1", 1000)
	id2 := createInflightTxnIssue39(t, client, refPrefix+"-2", 2000)

	result, resp, err := client.Transaction.BulkVoidInflight(blnkgo.BulkVoidInflightRequest{
		TransactionIDs: []string{id1, id2},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, 2, result.Succeeded)
	require.Equal(t, 0, result.Failed)
	require.Len(t, result.Results, 2)
	require.Equal(t, "succeeded", result.Results[0].Status)
	require.Equal(t, id1, result.Results[0].TransactionID)
	require.Equal(t, "succeeded", result.Results[1].Status)
	require.Equal(t, id2, result.Results[1].TransactionID)
}

func TestIssue39_BulkVoidInflight_EmptyList(t *testing.T) {
	client := newClientIssue39(t)

	result, resp, err := client.Transaction.BulkVoidInflight(blnkgo.BulkVoidInflightRequest{})
	require.Error(t, err)
	require.Nil(t, result)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "transaction_ids array cannot be empty")
}
