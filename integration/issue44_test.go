//go:build integration

// Integration tests for issue #44 — Reconciliation.DeleteMatchingRule (DELETE /reconciliation/matching-rules/{rule_id}).
// Requires Blnk Core running at http://localhost:5001 (docker compose up in blnk/).
//
// Run: go test -tags=integration -v ./integration/... -run Issue44
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue44_DeleteMatchingRule(t *testing.T) {
	client := newIntegrationClient(t)

	created, resp, err := client.Reconciliation.CreateMatchingRule(blnkgo.Matcher{
		Name:        fmt.Sprintf("Issue44 rule %d", time.Now().UnixNano()),
		Description: "Rule to delete",
		Criteria: []blnkgo.Criteria{
			{
				Field:    blnkgo.CriteriaFieldAmount,
				Operator: blnkgo.ReconciliationOperatorEquals,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, created.RuleID)

	deleted, resp, err := client.Reconciliation.DeleteMatchingRule(created.RuleID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotEmpty(t, deleted.Message)
}

func TestIssue44_DeleteMatchingRule_EmptyID(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Reconciliation.DeleteMatchingRule("")
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "matching rule id is required")
}

func TestIssue44_DeleteMatchingRule_NotFound(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Reconciliation.DeleteMatchingRule("rule_nonexistent_issue44")
	require.Error(t, err)
	require.NotNil(t, resp)
	require.NotEqual(t, http.StatusOK, resp.StatusCode)
}
