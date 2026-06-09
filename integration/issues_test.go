//go:build integration

// Integration tests for SDK methods added in issues #31–#35.
// Requires Blnk Core running at http://localhost:5001 (docker compose up in blnk/).
//
// Run: go test -tags=integration -v ./integration/...
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

const baseURL = "http://localhost:5001/"

func newClient(t *testing.T) *blnkgo.Client {
	t.Helper()
	u, err := url.Parse(baseURL)
	require.NoError(t, err)
	return blnkgo.NewClient(u, nil, blnkgo.WithTimeout(15*time.Second), blnkgo.WithRetry(2))
}

func postJSON(t *testing.T, path string, body any) (*http.Response, []byte) {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)
	resp, err := http.Post(baseURL+path, "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()
	return resp, data
}

func createIndividualIdentity(t *testing.T, client *blnkgo.Client, label string) *blnkgo.IdentityResponse {
	t.Helper()
	dob := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)
	identity, resp, err := client.Identity.Create(blnkgo.Identity{
		IdentityType: blnkgo.Individual,
		FirstName:    label,
		LastName:     "Integration",
		DOB:          &dob,
		Gender:       "male",
		Nationality:  "Nigerian",
		EmailAddress: fmt.Sprintf("%s-%d@test.com", label, time.Now().UnixNano()),
		PhoneNumber:  "+2348012345678",
		Category:     "customer",
		Street:       "1 Test St",
		Country:      "NG",
		State:        "Lagos",
		PostCode:     "100001",
		City:         "Lagos",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	return identity
}

// Issue #31 — Ledger.Update (PUT /ledgers/{id})
func TestIssue31_LedgerUpdate(t *testing.T) {
	client := newClient(t)

	ledger, resp, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{
		Name: "SDK Integration Ledger",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	updated, resp, err := client.Ledger.Update(ledger.LedgerID, blnkgo.UpdateLedgerRequest{
		Name: "SDK Integration Ledger Updated",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "SDK Integration Ledger Updated", updated.Name)

	fetched, resp, err := client.Ledger.Get(ledger.LedgerID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "SDK Integration Ledger Updated", fetched.Name)
}

// Issue #32 — LedgerBalance.GetLineage (GET /balances/{id}/lineage)
func TestIssue32_BalanceLineage(t *testing.T) {
	client := newClient(t)

	ledger, _, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{Name: "Lineage Ledger"})
	require.NoError(t, err)

	identity := createIndividualIdentity(t, client, "lineage")

	resp, raw := postJSON(t, "balances", map[string]any{
		"ledger_id":          ledger.LedgerID,
		"identity_id":        identity.IdentityId,
		"currency":           "USD",
		"track_fund_lineage": true,
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var balance struct {
		BalanceID string `json:"balance_id"`
	}
	require.NoError(t, json.Unmarshal(raw, &balance))

	lineage, resp, err := client.LedgerBalance.GetLineage(balance.BalanceID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, balance.BalanceID, lineage.BalanceID)
	require.NotNil(t, lineage.Providers)
}

// Issue #33 — LedgerBalance.UpdateIdentity (PUT /balances/{id}/identity)
func TestIssue33_UpdateBalanceIdentity(t *testing.T) {
	client := newClient(t)

	ledger, _, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{Name: "Identity Update Ledger"})
	require.NoError(t, err)

	balance, resp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID: ledger.LedgerID,
		Currency: "USD",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Empty(t, balance.IdentityID)

	identity := createIndividualIdentity(t, client, "owner")

	result, resp, err := client.LedgerBalance.UpdateIdentity(balance.BalanceID, blnkgo.UpdateBalanceIdentityRequest{
		IdentityID: identity.IdentityId,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, result.Message, "successfully")

	updated, resp, err := client.LedgerBalance.Get(balance.BalanceID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, identity.IdentityId, updated.IdentityID)
}

// Issue #34 — LedgerBalance.CreateSnapshot (POST /balances-snapshots)
func TestIssue34_CreateBalanceSnapshot(t *testing.T) {
	client := newClient(t)

	ledger, _, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{Name: "Snapshot Ledger"})
	require.NoError(t, err)

	_, resp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID: ledger.LedgerID,
		Currency: "USD",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	withBatch, resp, err := client.LedgerBalance.CreateSnapshot(blnkgo.CreateBalanceSnapshotRequest{
		BatchSize: 500,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, withBatch.Message, "Snapshotting in progress")

	defaultBatch, resp, err := client.LedgerBalance.CreateSnapshot(blnkgo.CreateBalanceSnapshotRequest{})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, defaultBatch.Message, "Snapshotting in progress")
}

// Issue #34 — SDK client-side validation for negative batch_size
func TestIssue34_CreateBalanceSnapshot_InvalidBatchSize(t *testing.T) {
	client := newClient(t)

	_, resp, err := client.LedgerBalance.CreateSnapshot(blnkgo.CreateBalanceSnapshotRequest{
		BatchSize: -1,
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "batch_size must be positive")
}

// Issue #35 — Transaction.CreateBulk (POST /transactions/bulk)
func TestIssue35_CreateBulkTransactions(t *testing.T) {
	client := newClient(t)
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

// Issue #35 — SDK client-side validation for empty bulk request
func TestIssue35_CreateBulkTransactions_EmptyList(t *testing.T) {
	client := newClient(t)

	_, resp, err := client.Transaction.CreateBulk(blnkgo.CreateBulkTransactionRequest{})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "transactions array cannot be empty")
}
