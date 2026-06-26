//go:build integration

// Integration tests for issue #38 — Transaction.BulkCommitInflight (POST /transactions/inflight/bulk/commit).
// Requires Blnk Core running at http://localhost:5001 (docker compose up in blnk/).
//
// Run: go test -tags=integration -v ./integration/... -run Issue38
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func newClientIssue38(t *testing.T) *blnkgo.Client {
	return newIntegrationClient(t)
}

func createInflightTxnIssue38(t *testing.T, client *blnkgo.Client, ref string, amount float64) string {
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
			Description: "Bulk commit inflight setup",
		},
		Inflight:           true,
		InflightExpiryDate: &expiry,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, txn.TransactionID)
	return txn.TransactionID
}

func TestIssue38_BulkCommitInflight(t *testing.T) {
	client := newClientIssue38(t)
	refPrefix := fmt.Sprintf("bulk-commit-%d", time.Now().UnixNano())

	id1 := createInflightTxnIssue38(t, client, refPrefix+"-1", 1000)
	id2 := createInflightTxnIssue38(t, client, refPrefix+"-2", 2000)

	result, resp, err := client.Transaction.BulkCommitInflight(blnkgo.BulkCommitInflightRequest{
		Transactions: []blnkgo.BulkCommitInflightItem{
			{TransactionID: id1},
			{TransactionID: id2},
		},
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

func TestIssue38_BulkCommitInflight_EmptyList(t *testing.T) {
	client := newClientIssue38(t)

	result, resp, err := client.Transaction.BulkCommitInflight(blnkgo.BulkCommitInflightRequest{})
	require.Error(t, err)
	require.Nil(t, result)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "transactions array cannot be empty")
}
