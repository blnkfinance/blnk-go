//go:build integration

// Integration tests for issue #118 — BalanceMonitor.Delete (DELETE /balance-monitors/{id}).
// Requires Blnk Core 0.15.0+ at http://localhost:5001.
//
// Run: go test -tags=integration -v ./integration/... -run Issue118
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func createBalanceMonitorForTest(t *testing.T, client *blnkgo.Client) string {
	t.Helper()

	ledger, ledgerResp, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{
		Name: fmt.Sprintf("issue118-ledger-%d", time.Now().UnixNano()),
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, ledgerResp.StatusCode)

	balance, balanceResp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID: ledger.LedgerID,
		Currency: "USD",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, balanceResp.StatusCode)

	monitor, monitorResp, err := client.BalanceMonitor.Create(blnkgo.MonitorData{
		BalanceID: balance.BalanceID,
		Condition: blnkgo.MonitorCondition{
			Field:     "credit_balance",
			Operator:  blnkgo.OperatorGreaterThan,
			Value:     1000,
			Precision: 100,
		},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, monitorResp.StatusCode)
	require.NotEmpty(t, monitor.MonitorID)

	return monitor.MonitorID
}

func TestIssue118_DeleteBalanceMonitor(t *testing.T) {
	client := newIntegrationClient(t)
	monitorID := createBalanceMonitorForTest(t, client)

	deleted, deleteResp, err := client.BalanceMonitor.Delete(monitorID)
	require.NoError(t, err)
	require.NotNil(t, deleteResp)
	require.Equal(t, http.StatusOK, deleteResp.StatusCode)
	require.Equal(t, "BalanceMonitor deleted successfully", deleted.Message)

	_, getResp, err := client.BalanceMonitor.Get(monitorID)
	require.Error(t, err)
	require.NotNil(t, getResp)
	require.NotEqual(t, http.StatusOK, getResp.StatusCode)
}

func TestIssue118_DeleteBalanceMonitor_ValidationError(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.BalanceMonitor.Delete("")
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "monitor id is required")
}

func TestIssue118_DeleteBalanceMonitor_NotFound(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.BalanceMonitor.Delete("mon_nonexistent_issue118")
	require.Error(t, err)
	require.NotNil(t, resp)
	require.NotEqual(t, http.StatusOK, resp.StatusCode)
}
