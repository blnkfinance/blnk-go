//go:build integration

// Integration tests for issue #43 — Reconciliation.UpdateMatchingRule (PUT /reconciliation/matching-rules/{rule_id}).
// Requires Blnk Core running at http://localhost:5001 (docker compose up in blnk/).
//
// Run: go test -tags=integration -v ./integration/... -run Issue43
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue43_UpdateMatchingRule(t *testing.T) {
	client := newIntegrationClient(t)

	created, resp, err := client.Reconciliation.CreateMatchingRule(blnkgo.Matcher{
		Name:        fmt.Sprintf("Issue43 rule %d", time.Now().UnixNano()),
		Description: "Original description",
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

	updatedName := fmt.Sprintf("Issue43 updated %d", time.Now().UnixNano())
	updated, resp, err := client.Reconciliation.UpdateMatchingRule(created.RuleID, blnkgo.Matcher{
		Name:        updatedName,
		Description: "Updated description",
		Criteria: []blnkgo.Criteria{
			{
				Field:    blnkgo.CriteriaFieldReference,
				Operator: blnkgo.ReconciliationOperatorEquals,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, created.RuleID, updated.RuleID)
	require.Equal(t, updatedName, updated.Name)
	require.Equal(t, "Updated description", updated.Description)
	require.Len(t, updated.Criteria, 1)
	require.Equal(t, blnkgo.CriteriaFieldReference, updated.Criteria[0].Field)
}

func TestIssue43_UpdateMatchingRule_EmptyID(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Reconciliation.UpdateMatchingRule("", blnkgo.Matcher{Name: "x"})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "matching rule id is required")
}

func TestIssue43_UpdateMatchingRule_NotFound(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Reconciliation.UpdateMatchingRule("rule_nonexistent_issue43", blnkgo.Matcher{
		Name: "Updated",
		Criteria: []blnkgo.Criteria{
			{Field: blnkgo.CriteriaFieldAmount, Operator: blnkgo.ReconciliationOperatorEquals},
		},
	})
	require.Error(t, err)
	require.NotNil(t, resp)
	require.NotEqual(t, http.StatusOK, resp.StatusCode)
}
