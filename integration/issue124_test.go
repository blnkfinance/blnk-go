//go:build integration

// Integration tests for issue #124 — BalanceMonitor.ListByBalanceID
// (GET /balance-monitors/balances/{balance_id}).
// Requires Blnk Core running at http://localhost:5001 and BLNK_API_KEY in the environment.
//
// Run:
//
//	export BLNK_API_KEY=your_key
//	go test -tags=integration -v ./integration/... -run Issue124
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue124_ListByBalanceID(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	ledger, resp, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{
		Name: "Issue124 ListByBalance " + suffix,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	balance, resp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID: ledger.LedgerID,
		Currency: "USD",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	created, resp, err := client.BalanceMonitor.Create(blnkgo.MonitorData{
		BalanceID: balance.BalanceID,
		Condition: blnkgo.MonitorCondition{
			Field:     "credit_balance",
			Operator:  blnkgo.OperatorGreaterThan,
			Value:     1000,
			Precision: 100,
		},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, created.MonitorID)

	monitors, resp, err := client.BalanceMonitor.ListByBalanceID(balance.BalanceID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotEmpty(t, monitors)

	found := false
	for _, mon := range monitors {
		if mon.MonitorID == created.MonitorID {
			found = true
			require.Equal(t, balance.BalanceID, mon.BalanceID)
			break
		}
	}
	require.True(t, found, "created monitor should be returned for its balance")
}

func TestIssue124_ListByBalanceID_EmptyBalanceID(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.BalanceMonitor.ListByBalanceID("")
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "balance id is required")
}
