//go:build integration

// Integration tests for issue #35 — Transaction.CreateBulk (POST /transactions/bulk).
// Requires Blnk Core running at http://localhost:5001 (docker compose up in blnk/).
//
// Run: go test -tags=integration -v ./integration/... -run Issue35
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func newClientIssue35(t *testing.T) *blnkgo.Client {
	return newIntegrationClient(t)
}

func TestIssue35_CreateBulkTransactions(t *testing.T) {
	client := newClientIssue35(t)
	refPrefix := fmt.Sprintf("bulk-%d", time.Now().UnixNano())

	result, resp, err := client.Transaction.CreateBulk(blnkgo.CreateBulkTransactionRequest{
		Atomic: true,
		Transactions: []blnkgo.CreateTransactionRequest{
			{
				ParentTransaction: blnkgo.ParentTransaction{
					Amount:      500,
					Reference:   refPrefix + "-1",
					Precision:   100,
					Currency:    "USD",
					Source:      "@FundingPool",
					Destination: "@Recipient",
					Description: "Bulk integration tx 1",
				},
			},
			{
				ParentTransaction: blnkgo.ParentTransaction{
					Amount:      750,
					Reference:   refPrefix + "-2",
					Precision:   100,
					Currency:    "USD",
					Source:      "@FundingPool",
					Destination: "@Recipient",
					Description: "Bulk integration tx 2",
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, 2, result.TransactionCount)
	require.NotEmpty(t, result.BatchID)
	require.Equal(t, "applied", result.Status)
}

func TestIssue35_CreateBulkTransactions_EmptyList(t *testing.T) {
	client := newClientIssue35(t)

	_, resp, err := client.Transaction.CreateBulk(blnkgo.CreateBulkTransactionRequest{})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "transactions array cannot be empty")
}
