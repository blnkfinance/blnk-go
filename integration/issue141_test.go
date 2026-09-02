//go:build integration

// Integration tests for issue #141 — Core 0.15.3 SDK patch.
// Requires Blnk Core 0.15.3+ at http://localhost:5001 and BLNK_API_KEY.
//
// Run: go test -tags=integration -v ./integration/... -run Issue141
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue141_Core0153Patch(t *testing.T) {
	client := newIntegrationClient(t)

	health, healthResp, err := client.Health.Check()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, healthResp.StatusCode)
	require.Equal(t, "UP", health.Status)

	indicator := fmt.Sprintf("@SdkGo%d", time.Now().UnixNano())
	dest := fmt.Sprintf("@SdkGoDest%d", time.Now().UnixNano())

	gl, glResp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID:  blnkgo.GeneralLedgerID,
		Currency:  "USD",
		Indicator: indicator,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, glResp.StatusCode)
	require.Equal(t, indicator, gl.Indicator)
	require.Equal(t, blnkgo.GeneralLedgerID, gl.LedgerID)

	preview, previewResp, err := client.Transaction.CreateDryRun(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      25,
			Precision:   100,
			Reference:   fmt.Sprintf("dryrun-%d", time.Now().UnixNano()),
			Description: "SDK dry-run",
			Currency:    "USD",
			Source:      indicator,
			Destination: dest,
		},
		AllowOverdraft: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, previewResp.StatusCode)
	require.True(t, preview.DryRun)
	require.True(t, preview.WouldApply)
	require.NotEmpty(t, preview.Balances)
	require.Empty(t, preview.Rejection)

	bulkPreview, bulkResp, err := client.Transaction.CreateBulkDryRun(blnkgo.CreateBulkTransactionRequest{
		Transactions: []blnkgo.CreateTransactionRequest{
			{
				ParentTransaction: blnkgo.ParentTransaction{
					Amount:      10,
					Precision:   100,
					Reference:   fmt.Sprintf("bulk-a-%d", time.Now().UnixNano()),
					Description: "bulk preview 1",
					Currency:    "USD",
					Source:      indicator,
					Destination: dest,
				},
				AllowOverdraft: true,
			},
			{
				ParentTransaction: blnkgo.ParentTransaction{
					Amount:      15,
					Precision:   100,
					Reference:   fmt.Sprintf("bulk-b-%d", time.Now().UnixNano()),
					Description: "bulk preview 2",
					Currency:    "USD",
					Source:      indicator,
					Destination: dest,
				},
				AllowOverdraft: true,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, bulkResp.StatusCode)
	require.True(t, bulkPreview.DryRun)
	require.True(t, bulkPreview.WouldApply)

	posted, postedResp, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      20,
			Precision:   100,
			Reference:   fmt.Sprintf("posted-%d", time.Now().UnixNano()),
			Description: "SDK posted for refund",
			Currency:    "USD",
			Source:      indicator,
			Destination: dest,
			SkipQueue:   true,
		},
		AllowOverdraft: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, postedResp.StatusCode)
	require.NotEmpty(t, posted.TransactionID)

	refundPreview, refundPreviewResp, err := client.Transaction.RefundDryRun(posted.TransactionID, &blnkgo.RefundTransactionRequest{
		Description: "preview refund",
		MetaData:    blnkgo.MetaData{"channel": "sdk-test"},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, refundPreviewResp.StatusCode)
	require.True(t, refundPreview.DryRun)

	refund, refundResp, err := client.Transaction.Refund(posted.TransactionID, &blnkgo.RefundTransactionRequest{
		SkipQueue:   true,
		Description: "customer refund",
		MetaData:    blnkgo.MetaData{"channel": "sdk-test"},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, refundResp.StatusCode)
	require.NotEmpty(t, refund.TransactionID)
	require.Equal(t, "customer refund", refund.Description)

	hold, holdResp, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      30,
			Precision:   100,
			Reference:   fmt.Sprintf("hold-%d", time.Now().UnixNano()),
			Description: "SDK inflight hold",
			Currency:    "USD",
			Source:      indicator,
			Destination: dest,
			SkipQueue:   true,
		},
		AllowOverdraft: true,
		Inflight:       true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, holdResp.StatusCode)
	require.NotEmpty(t, hold.TransactionID)

	inflightPreview, inflightResp, err := client.Transaction.UpdateDryRun(hold.TransactionID, blnkgo.UpdateStatus{
		Status: blnkgo.InflightStatusCommit,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, inflightResp.StatusCode)
	require.True(t, inflightPreview.DryRun)
	require.Equal(t, "commit", inflightPreview.Operation)

	stillHold, _, err := client.Transaction.Get(hold.TransactionID)
	require.NoError(t, err)
	require.Equal(t, blnkgo.PryTransactionStatusInFlight, stillHold.Status)

	commitPreview, commitResp, err := client.Transaction.BulkCommitInflightDryRun(blnkgo.BulkCommitInflightRequest{
		SkipQueue:    true,
		Transactions: []blnkgo.BulkCommitInflightItem{{TransactionID: hold.TransactionID}},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, commitResp.StatusCode)
	require.True(t, commitPreview.DryRun)
	require.True(t, commitPreview.WouldApply)
	require.False(t, commitPreview.Cumulative)

	afterCommitPreview, _, err := client.Transaction.Get(hold.TransactionID)
	require.NoError(t, err)
	require.Equal(t, blnkgo.PryTransactionStatusInFlight, afterCommitPreview.Status)

	voidPreview, voidResp, err := client.Transaction.BulkVoidInflightDryRun(blnkgo.BulkVoidInflightRequest{
		SkipQueue:      true,
		TransactionIDs: []string{hold.TransactionID},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, voidResp.StatusCode)
	require.True(t, voidPreview.DryRun)

	afterVoidPreview, _, err := client.Transaction.Get(hold.TransactionID)
	require.NoError(t, err)
	require.Equal(t, blnkgo.PryTransactionStatusInFlight, afterVoidPreview.Status)

	hooks, hooksResp, err := client.Hooks.List(nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, hooksResp.StatusCode)
	require.NotNil(t, hooks)

	same, sameResp, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      10,
			Precision:   100,
			Reference:   fmt.Sprintf("same-%d", time.Now().UnixNano()),
			Description: "same source dest",
			Currency:    "USD",
			Source:      indicator,
			Destination: indicator,
			SkipQueue:   true,
		},
		AllowOverdraft: true,
	})
	require.Error(t, err)
	require.Nil(t, same)
	apiErr, ok := blnkgo.AsApiErrorResponse(err)
	require.True(t, ok)
	require.NotNil(t, sameResp)
	require.NotEqual(t, http.StatusCreated, sameResp.StatusCode)
	require.NotNil(t, apiErr.ErrorDetail)
	require.NotEmpty(t, apiErr.ErrorDetail.Code)

	dup, dupResp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID:  blnkgo.GeneralLedgerID,
		Currency:  "USD",
		Indicator: indicator,
	})
	require.Error(t, err)
	require.Nil(t, dup)
	require.Equal(t, http.StatusConflict, dupResp.StatusCode)
	dupErr, ok := blnkgo.AsApiErrorResponse(err)
	require.True(t, ok)
	require.NotNil(t, dupErr.ErrorDetail)
	require.Equal(t, blnkgo.ErrorCodeGenConflict, dupErr.ErrorDetail.Code)
}

func TestIssue141_CreateRejectsDryRunFlag(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      10,
			Precision:   100,
			Reference:   fmt.Sprintf("flag-%d", time.Now().UnixNano()),
			Description: "should not post",
			Currency:    "USD",
			Source:      "@World",
			Destination: "@Bank",
		},
		DryRun: true,
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "CreateDryRun")
}
