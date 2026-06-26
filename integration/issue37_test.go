//go:build integration

// Integration tests for issue #37 — Transaction.GetByReference (GET /transactions/reference/{reference}).
// Requires Blnk Core running at http://localhost:5001 (docker compose up in blnk/).
//
// Run: go test -tags=integration -v ./integration/... -run Issue37
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func newClientIssue37(t *testing.T) *blnkgo.Client {
	return newIntegrationClient(t)
}

func TestIssue37_GetTransactionByReference(t *testing.T) {
	client := newClientIssue37(t)
	ref := fmt.Sprintf("ref-by-ref-%d", time.Now().UnixNano())

	txn, resp, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      500,
			Reference:   ref,
			Precision:   100,
			Currency:    "USD",
			Source:      "@FundingPool",
			Destination: "@Recipient",
			Description: "GetByReference integration tx",
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
			require.NoError(t, getErr, "transaction not available before GetByReference call")
		}
		time.Sleep(200 * time.Millisecond)
	}

	byRef, resp, err := client.Transaction.GetByReference(ref)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, txn.TransactionID, byRef.TransactionID)
	require.Equal(t, ref, byRef.Reference)
	require.Equal(t, "USD", byRef.Currency)
}

func TestIssue37_GetTransactionByReference_EmptyReference(t *testing.T) {
	client := newClientIssue37(t)

	transaction, resp, err := client.Transaction.GetByReference("")
	require.Error(t, err)
	require.Nil(t, transaction)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "reference is required")
}
