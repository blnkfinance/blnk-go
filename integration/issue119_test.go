//go:build integration

// Integration tests for issue #119 — Reconciliation.Run response (POST /reconciliation/start).
// Requires Blnk Core 0.15.0+ at http://localhost:5001.
//
// Run: go test -tags=integration -v ./integration/... -run Issue119
package integration

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func createReconMatchingRuleForIssue119(t *testing.T, client *blnkgo.Client) string {
	t.Helper()

	matcher, resp, err := client.Reconciliation.CreateMatchingRule(blnkgo.Matcher{
		Name:        fmt.Sprintf("Issue119 rule %d", time.Now().UnixNano()),
		Description: "Amount match for reconciliation start",
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

func writeIssue119CSV(t *testing.T) string {
	t.Helper()

	extID := fmt.Sprintf("ext-issue119-%d", time.Now().UnixNano())
	content := fmt.Sprintf("ID,Amount,Currency,Reference,Description,Date\n%s,100.50,USD,ref-issue119,test row,2024-01-01T10:00:00Z\n", extID)
	path := filepath.Join(t.TempDir(), "issue119.csv")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestIssue119_Run_StartResponse(t *testing.T) {
	client := newIntegrationClient(t)
	ruleID := createReconMatchingRuleForIssue119(t, client)

	csvPath := writeIssue119CSV(t)
	uploaded, uploadResp, err := client.Reconciliation.Upload("integration-test", csvPath, "issue119.csv")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, uploadResp.StatusCode)
	require.NotEmpty(t, uploaded.UploadID)

	result, resp, err := client.Reconciliation.Run(blnkgo.RunReconData{
		UploadID:        uploaded.UploadID,
		Strategy:        blnkgo.ReconciliationStrategyOneToOne,
		DryRun:          true,
		MatchingRuleIDs: []string{ruleID},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotEmpty(t, result.ReconciliationID)
	require.Contains(t, result.ReconciliationID, "recon_")
}
