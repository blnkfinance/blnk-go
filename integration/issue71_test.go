//go:build integration

// Integration tests for issue #71 — precise_distribution on split legs (POST /transactions).
// Requires Blnk Core running at http://localhost:5001 and BLNK_API_KEY in the environment.
//
// Run:
//
//	export BLNK_API_KEY=your_key
//	go test -tags=integration -v ./integration/... -run Issue71
package integration

import (
	"fmt"
	"math/big"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue71_PreciseDistributionLegs(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	ledger, resp, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{Name: "Issue71 Precise " + suffix})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	merchant, resp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID: ledger.LedgerID,
		Currency: "USD",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	fee, resp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID: ledger.LedgerID,
		Currency: "USD",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	txn, resp, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			PreciseAmount: big.NewInt(10000),
			Reference:     "issue71-precise-" + suffix,
			Precision:     100,
			Currency:      "USD",
			Source:        "@FundingPool",
			Description:   "Issue 71 precise_distribution legs",
			SkipQueue:     true,
			Destinations: []blnkgo.Source{
				{Identifier: merchant.BalanceID, PreciseDistribution: "9733"},
				{Identifier: fee.BalanceID, PreciseDistribution: "267"},
			},
		},
		AllowOverdraft: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, txn.TransactionID)

	merchantAfter, resp, err := client.LedgerBalance.Get(merchant.BalanceID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, merchantAfter.CreditBalance)
	require.Equal(t, 0, big.NewInt(9733).Cmp(merchantAfter.CreditBalance),
		"merchant credit should equal precise_distribution leg amount")

	feeAfter, resp, err := client.LedgerBalance.Get(fee.BalanceID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, feeAfter.CreditBalance)
	require.Equal(t, 0, big.NewInt(267).Cmp(feeAfter.CreditBalance),
		"fee credit should equal precise_distribution leg amount")
}

func TestIssue71_MixedDecimalPreciseDistributionAndLeft(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	ledger, resp, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{Name: "Issue71 Mixed " + suffix})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	balances := make([]string, 3)
	for i := range balances {
		bal, resp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
			LedgerID: ledger.LedgerID,
			Currency: "USD",
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		balances[i] = bal.BalanceID
	}

	txn, resp, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      30000,
			Reference:   "issue71-mix-" + suffix,
			Precision:   100,
			Currency:    "USD",
			Source:      "@FundingPool",
			Description: "Issue 71 mixed split legs",
			SkipQueue:   true,
			Destinations: []blnkgo.Source{
				{Identifier: balances[0], Distribution: "33.33%"},
				{Identifier: balances[1], PreciseDistribution: "5000"},
				{Identifier: balances[2], Distribution: "left"},
			},
		},
		AllowOverdraft: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, txn.TransactionID)
}

func TestIssue71_PreciseAmountWithClassicDistribution_BackwardCompat(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	ledger, resp, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{Name: "Issue71 Compat " + suffix})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	destA, resp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID: ledger.LedgerID,
		Currency: "USD",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	destB, resp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID: ledger.LedgerID,
		Currency: "USD",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	txn, resp, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			PreciseAmount: big.NewInt(50000),
			Reference:     "issue71-compat-" + suffix,
			Precision:     100,
			Currency:      "USD",
			Source:        "@FundingPool",
			Description:   "Issue 71 precise_amount classic distribution compat",
			SkipQueue:     true,
			Destinations: []blnkgo.Source{
				{Identifier: destA.BalanceID, Distribution: "50%"},
				{Identifier: destB.BalanceID, Distribution: "left"},
			},
		},
		AllowOverdraft: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, txn.TransactionID)
}

func TestIssue71_PreciseAmountWithPreciseDistributionLegs(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	ledger, resp, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{Name: "Issue71 PreciseAmt " + suffix})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	dest, resp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID: ledger.LedgerID,
		Currency: "USD",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	txn, resp, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			PreciseAmount: big.NewInt(10000),
			Reference:     "issue71-pamt-" + suffix,
			Precision:     100,
			Currency:      "USD",
			Source:        "@FundingPool",
			Description:   "Issue 71 precise_amount split",
			SkipQueue:     true,
			Destinations: []blnkgo.Source{
				{Identifier: dest.BalanceID, PreciseDistribution: "9733"},
				{Identifier: "@Recipient", PreciseDistribution: "267"},
			},
		},
		AllowOverdraft: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, txn.TransactionID)
}
