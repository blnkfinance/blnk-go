package blnkgo_test

import (
	"math/big"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func baseSplitTxn() blnkgo.CreateTransactionRequest {
	return blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Precision:   100,
			Reference:   "ref_split_001",
			Description: "Split transaction",
			Currency:    "USD",
			Source:      "bln_source",
		},
	}
}

func TestValidateCreateTransaction_PreciseDistributionOnly(t *testing.T) {
	txn := baseSplitTxn()
	txn.PreciseAmount = big.NewInt(10000)
	txn.Destinations = []blnkgo.Source{
		{Identifier: "bln_merchant", PreciseDistribution: "9733"},
		{Identifier: "bln_fee", PreciseDistribution: "267"},
	}
	require.NoError(t, blnkgo.ValidateCreateTransacation(txn))
}

func TestValidateCreateTransaction_PreciseDistributionWithAmount(t *testing.T) {
	txn := baseSplitTxn()
	txn.Amount = 10000
	txn.Destinations = []blnkgo.Source{
		{Identifier: "bln_merchant", PreciseDistribution: "9733"},
		{Identifier: "bln_fee", PreciseDistribution: "267"},
	}
	require.NoError(t, blnkgo.ValidateCreateTransacation(txn))
}

func TestValidateCreateTransaction_MixedDecimalAndPreciseDistribution(t *testing.T) {
	txn := baseSplitTxn()
	txn.Amount = 30000
	txn.Destinations = []blnkgo.Source{
		{Identifier: "bln_a", Distribution: "33.33%"},
		{Identifier: "bln_b", PreciseDistribution: "5000"},
		{Identifier: "bln_c", Distribution: "left"},
	}
	require.NoError(t, blnkgo.ValidateCreateTransacation(txn))
}

func TestValidateCreateTransaction_PreciseDistributionSumMismatch(t *testing.T) {
	txn := baseSplitTxn()
	txn.Amount = 10000
	txn.Destinations = []blnkgo.Source{
		{Identifier: "bln_merchant", PreciseDistribution: "9733"},
		{Identifier: "bln_fee", PreciseDistribution: "300"},
	}
	err := blnkgo.ValidateCreateTransacation(txn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not equal")
}

func TestValidateCreateTransaction_MissingDistributionFields(t *testing.T) {
	txn := baseSplitTxn()
	txn.Amount = 1000
	txn.Destinations = []blnkgo.Source{
		{Identifier: "bln_merchant"},
	}
	err := blnkgo.ValidateCreateTransacation(txn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "precise_distribution")
}

func TestValidateCreateTransaction_InvalidPreciseDistribution(t *testing.T) {
	txn := baseSplitTxn()
	txn.Amount = 1000
	txn.Destinations = []blnkgo.Source{
		{Identifier: "bln_alice", PreciseDistribution: "not-a-number"},
	}
	err := blnkgo.ValidateCreateTransacation(txn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid precise_distribution")
}

func TestValidateCreateTransaction_LargePreciseDistributionExactSum(t *testing.T) {
	legA := "9007199254740992"
	legB := "1"
	total := "9007199254740993"

	txn := baseSplitTxn()
	txn.PreciseAmount = new(big.Int)
	_, ok := txn.PreciseAmount.SetString(total, 10)
	require.True(t, ok)
	txn.Destinations = []blnkgo.Source{
		{Identifier: "bln_alice", PreciseDistribution: legA},
		{Identifier: "bln_bob", PreciseDistribution: legB},
	}
	require.NoError(t, blnkgo.ValidateCreateTransacation(txn))

	txn.Destinations[1].PreciseDistribution = "2"
	err := blnkgo.ValidateCreateTransacation(txn)
	require.Error(t, err)
}

func TestValidateCreateTransaction_ClassicDistributionStillWorks(t *testing.T) {
	txn := baseSplitTxn()
	txn.Amount = 1000
	txn.Destinations = []blnkgo.Source{
		{Identifier: "bln_fee", Distribution: "50%"},
		{Identifier: "bln_recipient", Distribution: "left"},
	}
	require.NoError(t, blnkgo.ValidateCreateTransacation(txn))
}
