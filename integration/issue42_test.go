//go:build integration

// Integration tests for issue #42 — Reconciliation.Get (GET /reconciliation/{reconciliation_id}).
// Requires Blnk Core running at http://localhost:5001 (docker compose up in blnk/).
//
// Run: go test -tags=integration -v ./integration/... -run Issue42
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue42_Get_AfterRunInstant(t *testing.T) {
	client := newIntegrationClient(t)
	ruleID := createInstantReconMatchingRule(t, client)

	extID := fmt.Sprintf("ext-get-%d", time.Now().UnixNano())
	started, resp, err := client.Reconciliation.RunInstant(blnkgo.RunInstantReconData{
		ExternalTransactions: []blnkgo.ExternalTransaction{
			{ID: extID, Amount: 1, Reference: "r1", Currency: "USD"},
		},
		Strategy:        blnkgo.ReconciliationStrategyOneToOne,
		DryRun:          true,
		MatchingRuleIDs: []string{ruleID},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotEmpty(t, started.ReconciliationID)

	recon, resp, err := client.Reconciliation.Get(started.ReconciliationID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, started.ReconciliationID, recon.ReconciliationID)
	require.NotEmpty(t, recon.Status)
	require.NotEmpty(t, recon.UploadID)
	require.False(t, recon.StartedAt.IsZero())
}

func TestIssue42_Get_EmptyID(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Reconciliation.Get("")
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "reconciliation id is required")
}

func TestIssue42_Get_NotFound(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Reconciliation.Get("recon_nonexistent_issue42")
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
