package blnkgo_test

import (
	"encoding/json"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestValidateCreateLedgerBalance_AllocationStrategy(t *testing.T) {
	require.NoError(t, blnkgo.ValidateCreateLedgerBalance(blnkgo.CreateLedgerBalanceRequest{
		LedgerID:           "ldg_1",
		Currency:           "USD",
		IdentityID:         "idt_1",
		TrackFundLineage:   true,
		AllocationStrategy: blnkgo.AllocationStrategyPROPORTIONAL,
	}))

	err := blnkgo.ValidateCreateLedgerBalance(blnkgo.CreateLedgerBalanceRequest{
		LedgerID:           "ldg_1",
		Currency:           "USD",
		AllocationStrategy: blnkgo.AllocationStrategy("INVALID"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "allocation_strategy must be one of FIFO, LIFO, or PROPORTIONAL")
}

func TestValidateCreateLedgerBalance_TrackFundLineageRequiresIdentity(t *testing.T) {
	err := blnkgo.ValidateCreateLedgerBalance(blnkgo.CreateLedgerBalanceRequest{
		LedgerID:         "ldg_1",
		Currency:         "USD",
		TrackFundLineage: true,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "identity_id is required when track_fund_lineage is enabled")
}

func TestCreateLedgerBalanceRequest_JSONMarshal(t *testing.T) {
	body := blnkgo.CreateLedgerBalanceRequest{
		LedgerID:           "ldg_1",
		IdentityID:         "idt_1",
		Currency:           "USD",
		TrackFundLineage:   true,
		AllocationStrategy: blnkgo.AllocationStrategyLIFO,
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(payload, &decoded))
	require.Equal(t, true, decoded["track_fund_lineage"])
	require.Equal(t, "LIFO", decoded["allocation_strategy"])
}

func TestLedgerBalance_UnmarshalJSON_LineageFields(t *testing.T) {
	payload := []byte(`{
		"balance_id": "bln_1",
		"ledger_id": "ldg_1",
		"currency": "USD",
		"track_fund_lineage": true,
		"allocation_strategy": "PROPORTIONAL"
	}`)

	var bal blnkgo.LedgerBalance
	require.NoError(t, json.Unmarshal(payload, &bal))
	require.True(t, bal.TrackFundLineage)
	require.Equal(t, blnkgo.AllocationStrategyPROPORTIONAL, bal.AllocationStrategy)
}
