package blnkgo_test

import (
	"fmt"
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

func TestValidateCreateTransaction_NonIntegerAmountWithPreciseDistributionRejected(t *testing.T) {
	txn := baseSplitTxn()
	txn.Amount = 10000.50
	txn.Destinations = []blnkgo.Source{
		{Identifier: "bln_merchant", PreciseDistribution: "9733"},
		{Identifier: "bln_fee", PreciseDistribution: "267"},
	}
	err := blnkgo.ValidateCreateTransacation(txn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "whole number")
}

func TestValidateCreateTransaction_NonIntegerAmountTruncationWouldMismatch(t *testing.T) {
	// 10000.99 truncates to 10000, which would incorrectly pass a 10001 split sum.
	txn := baseSplitTxn()
	txn.Amount = 10000.99
	txn.Destinations = []blnkgo.Source{
		{Identifier: "bln_merchant", PreciseDistribution: "9733"},
		{Identifier: "bln_fee", PreciseDistribution: "268"},
	}
	err := blnkgo.ValidateCreateTransacation(txn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "whole number")
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

func TestValidateCreateTransaction_PreciseAmountWithClassicDistributionSkipsSumValidation(t *testing.T) {
	// Pre-#71 behavior: when precise_amount is set without precise_distribution legs,
	// SDK does not client-side validate distribution sums; Core is source of truth.
	txn := baseSplitTxn()
	txn.PreciseAmount = big.NewInt(100000)
	txn.Destinations = []blnkgo.Source{
		{Identifier: "bln_fee", Distribution: "50%"},
		{Identifier: "bln_recipient", Distribution: "left"},
	}
	require.NoError(t, blnkgo.ValidateCreateTransacation(txn))
}

func TestValidateCreateTransaction_PreciseAmountWithPreciseDistributionStillValidated(t *testing.T) {
	txn := baseSplitTxn()
	txn.PreciseAmount = big.NewInt(10000)
	txn.Destinations = []blnkgo.Source{
		{Identifier: "bln_merchant", PreciseDistribution: "9733"},
		{Identifier: "bln_fee", PreciseDistribution: "300"},
	}
	err := blnkgo.ValidateCreateTransacation(txn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not equal")
}

func bulkCreateTransactions(n int) []blnkgo.CreateTransactionRequest {
	txns := make([]blnkgo.CreateTransactionRequest, n)
	for i := range txns {
		txns[i] = blnkgo.CreateTransactionRequest{
			ParentTransaction: blnkgo.ParentTransaction{
				Amount:      100,
				Reference:   fmt.Sprintf("bulk-ref-%d", i),
				Precision:   100,
				Currency:    "USD",
				Source:      "@FundingPool",
				Destination: "@Recipient",
				Description: "Bulk transaction",
			},
		}
	}
	return txns
}

func TestValidateCreateBulkTransaction_AtLimit(t *testing.T) {
	body := blnkgo.CreateBulkTransactionRequest{
		Transactions: bulkCreateTransactions(blnkgo.MaxBulkCreateItems),
	}
	require.NoError(t, blnkgo.ValidateCreateBulkTransaction(body))
}

func TestValidateCreateBulkTransaction_OverLimit(t *testing.T) {
	body := blnkgo.CreateBulkTransactionRequest{
		Transactions: bulkCreateTransactions(blnkgo.MaxBulkCreateItems + 1),
	}
	err := blnkgo.ValidateCreateBulkTransaction(body)
	require.Error(t, err)
	require.Contains(t, err.Error(), "too many transactions")
	require.Contains(t, err.Error(), "10000")
}
