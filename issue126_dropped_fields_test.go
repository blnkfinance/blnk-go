package blnkgo_test

import (
	"encoding/json"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Issue #126 — Core 0.15.0 dropped response fields (rate, currency_multiplier).
// modification_ref was never modeled on Go SDK types.

func TestIssue126_LedgerBalance_OmitsCurrencyMultiplier(t *testing.T) {
	payload := `{
		"balance_id": "bln_test",
		"ledger_id": "ldg_test",
		"currency": "USD",
		"precision": 100,
		"identity_id": "",
		"indicator": "",
		"created_at": "2026-07-14T00:00:00Z",
		"inflight_expires_at": "0001-01-01T00:00:00Z"
	}`

	var balance blnkgo.LedgerBalance
	require.NoError(t, json.Unmarshal([]byte(payload), &balance))
	assert.Nil(t, balance.CurrencyMultiplier)
}

func TestIssue126_LedgerBalance_LegacyCurrencyMultiplier(t *testing.T) {
	payload := `{
		"balance_id": "bln_test",
		"ledger_id": "ldg_test",
		"currency": "USD",
		"precision": 100,
		"currency_multiplier": 100,
		"identity_id": "",
		"indicator": "",
		"created_at": "2026-07-14T00:00:00Z",
		"inflight_expires_at": "0001-01-01T00:00:00Z"
	}`

	var balance blnkgo.LedgerBalance
	require.NoError(t, json.Unmarshal([]byte(payload), &balance))
	require.NotNil(t, balance.CurrencyMultiplier)
	assert.Equal(t, 100.0, *balance.CurrencyMultiplier)
}

func TestIssue126_Transaction_OmitsRate(t *testing.T) {
	payload := `{
		"transaction_id": "txn_test",
		"amount": 10,
		"precision": 100,
		"reference": "ref_001",
		"description": "test",
		"currency": "USD",
		"status": "APPLIED",
		"source": "@FundingPool",
		"destination": "@Dest",
		"created_at": "2026-07-14T00:00:00Z"
	}`

	var txn blnkgo.Transaction
	require.NoError(t, json.Unmarshal([]byte(payload), &txn))
	assert.Nil(t, txn.Rate)
}

func TestIssue126_SearchDocument_IgnoresLegacyRate(t *testing.T) {
	payload := `{
		"id": "txn_test",
		"transaction_id": "txn_test",
		"status": "APPLIED",
		"created_at": 1781028226,
		"rate": 1
	}`

	var doc blnkgo.SearchDocument
	require.NoError(t, json.Unmarshal([]byte(payload), &doc))
	assert.Equal(t, "txn_test", doc.TransactionID)
	assert.Equal(t, "APPLIED", doc.Status)
}
