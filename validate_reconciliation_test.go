package blnkgo_test

import (
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func validRunInstantReconData() blnkgo.RunInstantReconData {
	return blnkgo.RunInstantReconData{
		ExternalTransactions: []blnkgo.ExternalTransaction{
			{
				ID:          "ext-1",
				Amount:      5.49,
				Reference:   "INV-2023-002",
				Currency:    "GBP",
				Description: "Card payment",
				Date:        func() *time.Time { t := time.Date(2024, 11, 15, 14, 25, 30, 0, time.UTC); return &t }(),
				Source:      "bank-api",
			},
		},
		Strategy:        blnkgo.ReconciliationStrategyOneToOne,
		MatchingRuleIDs: []string{"rule_abc123"},
	}
}

func TestValidateRunInstantReconData_Valid(t *testing.T) {
	require.NoError(t, blnkgo.ValidateRunInstantReconData(validRunInstantReconData()))
}

func TestValidateRunInstantReconData_MinimalExternalTransaction(t *testing.T) {
	data := blnkgo.RunInstantReconData{
		ExternalTransactions: []blnkgo.ExternalTransaction{
			{ID: "ext-1", Amount: 1, Reference: "r1", Currency: "USD"},
		},
		Strategy:        blnkgo.ReconciliationStrategyOneToOne,
		MatchingRuleIDs: []string{"rule_abc123"},
	}
	require.NoError(t, blnkgo.ValidateRunInstantReconData(data))
}

func TestValidateRunInstantReconData_EmptyExternalTransactions(t *testing.T) {
	data := validRunInstantReconData()
	data.ExternalTransactions = nil
	err := blnkgo.ValidateRunInstantReconData(data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "external_transactions must be a non-empty array")
}

func TestValidateRunInstantReconData_EmptyStrategy(t *testing.T) {
	data := validRunInstantReconData()
	data.Strategy = ""
	err := blnkgo.ValidateRunInstantReconData(data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "strategy is required")
}

func TestValidateRunInstantReconData_EmptyMatchingRuleIDs(t *testing.T) {
	data := validRunInstantReconData()
	data.MatchingRuleIDs = nil
	err := blnkgo.ValidateRunInstantReconData(data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "matching_rule_ids must be a non-empty array")
}
