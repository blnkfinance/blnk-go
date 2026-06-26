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
				Date:        time.Date(2024, 11, 15, 14, 25, 30, 0, time.UTC),
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

func TestValidateRunInstantReconData_EmptyExternalTransactions(t *testing.T) {
	data := validRunInstantReconData()
	data.ExternalTransactions = nil
	err := blnkgo.ValidateRunInstantReconData(data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "external_transactions must be a non-empty array")
}

func TestValidateRunInstantReconData_InvalidStrategy(t *testing.T) {
	data := validRunInstantReconData()
	data.Strategy = "invalid"
	err := blnkgo.ValidateRunInstantReconData(data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "strategy must be one of")
}

func TestValidateRunInstantReconData_EmptyMatchingRuleIDs(t *testing.T) {
	data := validRunInstantReconData()
	data.MatchingRuleIDs = nil
	err := blnkgo.ValidateRunInstantReconData(data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "matching_rule_ids must be a non-empty array")
}

func TestValidateRunInstantReconData_MissingExternalTransactionField(t *testing.T) {
	data := validRunInstantReconData()
	data.ExternalTransactions[0].Reference = ""
	err := blnkgo.ValidateRunInstantReconData(data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "each external transaction must include")
}

func TestValidateRunInstantReconData_ZeroDate(t *testing.T) {
	data := validRunInstantReconData()
	data.ExternalTransactions[0].Date = time.Time{}
	err := blnkgo.ValidateRunInstantReconData(data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "each external transaction must include")
}
