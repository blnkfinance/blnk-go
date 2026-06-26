//go:build integration

// Integration tests for issue #41 — Reconciliation.RunInstant (POST /reconciliation/start-instant).
// Requires Blnk Core running at http://localhost:5001 (docker compose up in blnk/).
//
// Run: go test -tags=integration -v ./integration/... -run Issue41
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func createInstantReconMatchingRule(t *testing.T, client *blnkgo.Client) string {
	t.Helper()

	matcher, resp, err := client.Reconciliation.CreateMatchingRule(blnkgo.Matcher{
		Name:        fmt.Sprintf("Issue41 rule %d", time.Now().UnixNano()),
		Description: "Amount match for instant reconciliation",
		Criteria: []blnkgo.Criteria{
			{
				Field:    blnkgo.CriteriaFieldAmount,
				Operator: blnkgo.ReconciliationOperatorEquals,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, matcher.RuleID)
	return matcher.RuleID
}

func TestIssue41_RunInstant_DryRun(t *testing.T) {
	client := newIntegrationClient(t)
	ruleID := createInstantReconMatchingRule(t, client)

	extID := fmt.Sprintf("ext-%d", time.Now().UnixNano())
	txnDate := time.Now().UTC().Truncate(time.Second)
	result, resp, err := client.Reconciliation.RunInstant(blnkgo.RunInstantReconData{
		ExternalTransactions: []blnkgo.ExternalTransaction{
			{
				ID:          extID,
				Amount:      42.50,
				Reference:   "INV-41",
				Currency:    "USD",
				Description: "Instant recon test",
				Date:        &txnDate,
				Source:      "integration-test",
			},
		},
		Strategy:        blnkgo.ReconciliationStrategyOneToOne,
		DryRun:          true,
		MatchingRuleIDs: []string{ruleID},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotEmpty(t, result.ReconciliationID)
	require.Contains(t, result.ReconciliationID, "recon_")
}

func TestIssue41_RunInstant_MinimalPayload(t *testing.T) {
	client := newIntegrationClient(t)
	ruleID := createInstantReconMatchingRule(t, client)

	extID := fmt.Sprintf("ext-min-%d", time.Now().UnixNano())
	result, resp, err := client.Reconciliation.RunInstant(blnkgo.RunInstantReconData{
		ExternalTransactions: []blnkgo.ExternalTransaction{
			{
				ID:        extID,
				Amount:    1,
				Reference: "r1",
				Currency:  "USD",
			},
		},
		Strategy:        blnkgo.ReconciliationStrategyOneToOne,
		DryRun:          true,
		MatchingRuleIDs: []string{ruleID},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotEmpty(t, result.ReconciliationID)
}

func TestIssue41_RunInstant_ValidationError(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Reconciliation.RunInstant(blnkgo.RunInstantReconData{
		ExternalTransactions: nil,
		Strategy:             blnkgo.ReconciliationStrategyOneToOne,
		MatchingRuleIDs:      []string{"rule_placeholder"},
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "external_transactions must be a non-empty array")
}
