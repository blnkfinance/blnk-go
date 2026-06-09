//go:build integration

// Integration tests for issue #36 — Transaction.GetLineage (GET /transactions/{id}/lineage).
// Requires Blnk Core running at http://localhost:5001 (docker compose up in blnk/).
//
// Run: go test -tags=integration -v ./integration/... -run Issue36
package integration

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func newClientIssue36(t *testing.T) *blnkgo.Client {
	t.Helper()
	u, err := url.Parse("http://localhost:5001/")
	require.NoError(t, err)
	return blnkgo.NewClient(u, nil, blnkgo.WithTimeout(15*time.Second), blnkgo.WithRetry(2))
}

func TestIssue36_GetTransactionLineage(t *testing.T) {
	client := newClientIssue36(t)
	ref := fmt.Sprintf("lineage-%d", time.Now().UnixNano())

	txn, resp, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      500,
			Reference:   ref,
			Precision:   100,
			Currency:    "USD",
			Source:      "@FundingPool",
			Destination: "@Recipient",
			Description: "Lineage integration tx",
		},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, txn.TransactionID)

	deadline := time.Now().Add(15 * time.Second)
	for {
		_, getResp, getErr := client.Transaction.Get(txn.TransactionID)
		if getErr == nil && getResp.StatusCode == http.StatusOK {
			break
		}
		if time.Now().After(deadline) {
			require.NoError(t, getErr, "transaction not available before lineage call")
		}
		time.Sleep(200 * time.Millisecond)
	}

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
