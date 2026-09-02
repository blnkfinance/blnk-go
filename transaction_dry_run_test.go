package blnkgo_test

import (
	"encoding/json"
	"net/http"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateTransactionRequest_JSONIncludesDryRun(t *testing.T) {
	body := blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      25,
			Reference:   "ref_dry",
			Precision:   100,
			Currency:    "USD",
			Source:      "@src",
			Destination: "@dst",
			Description: "preview",
		},
		AllowOverdraft: true,
		DryRun:         true,
	}

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(payload, &decoded))
	assert.Equal(t, true, decoded["dry_run"])
}

func TestTransactionPreview_UnmarshalJSON(t *testing.T) {
	payload := []byte(`{
		"dry_run": true,
		"would_apply": false,
		"rejection": {"code": "TXN_INSUFFICIENT_FUNDS", "reason": "insufficient", "message": "insufficient funds"},
		"status": "REJECTED",
		"reference": "ref_1",
		"currency": "USD",
		"amount": 25,
		"precise_amount": "2500",
		"precision": 100,
		"operation": "commit",
		"balances": [
			{
				"balance_id": "@src",
				"role": "source",
				"currency": "USD",
				"virtual": true,
				"current_balance": "0",
				"resulting_balance": "0"
			}
		]
	}`)

	var preview blnkgo.TransactionPreview
	require.NoError(t, json.Unmarshal(payload, &preview))
	assert.True(t, preview.DryRun)
	assert.False(t, preview.WouldApply)
	require.NotNil(t, preview.Rejection)
	assert.Equal(t, "TXN_INSUFFICIENT_FUNDS", preview.Rejection.Code)
	assert.Equal(t, "commit", preview.Operation)
	assert.Equal(t, "2500", preview.PreciseAmount)
	require.Len(t, preview.Balances, 1)
	assert.Equal(t, "source", preview.Balances[0].Role)
}

func TestTransactionService_Create_RejectsDryRunFlag(t *testing.T) {
	mockClient, svc := setupTransactionService()

	_, resp, err := svc.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      10,
			Reference:   "ref",
			Source:      "@src",
			Destination: "@dst",
		},
		DryRun: true,
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CreateDryRun")
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
}

func TestTransactionService_CreateDryRun_Success(t *testing.T) {
	mockClient, svc := setupTransactionService()
	body := blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      25,
			Reference:   "ref_dry",
			Precision:   100,
			Currency:    "USD",
			Source:      "@src",
			Destination: "@dst",
			Description: "preview",
		},
		AllowOverdraft: true,
	}

	mockClient.On("NewRequest", "transactions", http.MethodPost, mock.MatchedBy(func(sent blnkgo.CreateTransactionRequest) bool {
		return sent.DryRun && sent.Reference == body.Reference
	})).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		preview := args.Get(1).(*blnkgo.TransactionPreview)
		*preview = blnkgo.TransactionPreview{
			DryRun:     true,
			WouldApply: true,
			Currency:   "USD",
			Amount:     25,
			Balances:   []blnkgo.BalanceProjection{{BalanceID: "@src", Role: "source"}},
		}
	})

	preview, resp, err := svc.CreateDryRun(body)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, preview.DryRun)
	assert.True(t, preview.WouldApply)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_CreateBulk_RejectsDryRunFlag(t *testing.T) {
	mockClient, svc := setupTransactionService()
	body := validBulkTransactionRequest()
	body.DryRun = true

	_, resp, err := svc.CreateBulk(body)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CreateBulkDryRun")
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
}

func TestTransactionService_CreateBulkDryRun_Success(t *testing.T) {
	mockClient, svc := setupTransactionService()
	body := validBulkTransactionRequest()

	mockClient.On("NewRequest", "transactions/bulk", http.MethodPost, mock.MatchedBy(func(sent blnkgo.CreateBulkTransactionRequest) bool {
		return sent.DryRun && len(sent.Transactions) == len(body.Transactions)
	})).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		preview := args.Get(1).(*blnkgo.BulkTransactionPreview)
		*preview = blnkgo.BulkTransactionPreview{DryRun: true, WouldApply: true, Cumulative: true}
	})

	preview, resp, err := svc.CreateBulkDryRun(body)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, preview.DryRun)
	assert.True(t, preview.WouldApply)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_Refund_RejectsDryRunFlag(t *testing.T) {
	mockClient, svc := setupTransactionService()

	_, resp, err := svc.Refund("txn-1", &blnkgo.RefundTransactionRequest{DryRun: true})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RefundDryRun")
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
}

func TestTransactionService_RefundDryRun_Success(t *testing.T) {
	mockClient, svc := setupTransactionService()

	mockClient.On("NewRequest", "refund-transaction/txn-1", http.MethodPost, mock.MatchedBy(func(sent *blnkgo.RefundTransactionRequest) bool {
		return sent != nil && sent.DryRun && sent.Description == "custom refund"
	})).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		preview := args.Get(1).(*blnkgo.TransactionPreview)
		*preview = blnkgo.TransactionPreview{DryRun: true, WouldApply: true}
	})

	preview, resp, err := svc.RefundDryRun("txn-1", &blnkgo.RefundTransactionRequest{
		Description: "custom refund",
		MetaData:    blnkgo.MetaData{"reason": "customer"},
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, preview.DryRun)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_Update_RejectsDryRunFlag(t *testing.T) {
	mockClient, svc := setupTransactionService()

	_, resp, err := svc.Update("txn-1", blnkgo.UpdateStatus{
		Status: blnkgo.InflightStatusCommit,
		DryRun: true,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UpdateDryRun")
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
}

func TestTransactionService_UpdateDryRun_Success(t *testing.T) {
	mockClient, svc := setupTransactionService()

	mockClient.On("NewRequest", "transactions/inflight/txn-1", http.MethodPut, mock.MatchedBy(func(sent blnkgo.UpdateStatus) bool {
		return sent.DryRun && sent.Status == blnkgo.InflightStatusCommit
	})).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		preview := args.Get(1).(*blnkgo.TransactionPreview)
		*preview = blnkgo.TransactionPreview{DryRun: true, WouldApply: true, Operation: "commit"}
	})

	preview, resp, err := svc.UpdateDryRun("txn-1", blnkgo.UpdateStatus{Status: blnkgo.InflightStatusCommit})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "commit", preview.Operation)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_BulkCommitInflight_RejectsDryRunFlag(t *testing.T) {
	mockClient, svc := setupTransactionService()
	body := validBulkCommitInflightRequest()
	body.DryRun = true

	_, resp, err := svc.BulkCommitInflight(body)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "BulkCommitInflightDryRun")
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
}

func TestTransactionService_BulkVoidInflight_RejectsDryRunFlag(t *testing.T) {
	mockClient, svc := setupTransactionService()

	_, resp, err := svc.BulkVoidInflight(blnkgo.BulkVoidInflightRequest{
		TransactionIDs: []string{"txn-1"},
		DryRun:         true,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "BulkVoidInflightDryRun")
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
}

func TestTransactionService_BulkCommitInflightDryRun_Success(t *testing.T) {
	mockClient, svc := setupTransactionService()
	body := validBulkCommitInflightRequest()

	mockClient.On("NewRequest", "transactions/inflight/bulk/commit", http.MethodPost, mock.MatchedBy(func(sent blnkgo.BulkCommitInflightRequest) bool {
		return sent.DryRun
	})).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		preview := args.Get(1).(*blnkgo.BulkTransactionPreview)
		*preview = blnkgo.BulkTransactionPreview{DryRun: true, WouldApply: true, Cumulative: false}
	})

	preview, resp, err := svc.BulkCommitInflightDryRun(body)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, preview.DryRun)
	assert.False(t, preview.Cumulative)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_BulkVoidInflightDryRun_Success(t *testing.T) {
	mockClient, svc := setupTransactionService()
	body := blnkgo.BulkVoidInflightRequest{TransactionIDs: []string{"txn-1"}}

	mockClient.On("NewRequest", "transactions/inflight/bulk/void", http.MethodPost, mock.MatchedBy(func(sent blnkgo.BulkVoidInflightRequest) bool {
		return sent.DryRun && len(sent.TransactionIDs) == 1
	})).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		preview := args.Get(1).(*blnkgo.BulkTransactionPreview)
		*preview = blnkgo.BulkTransactionPreview{DryRun: true, WouldApply: true}
	})

	preview, resp, err := svc.BulkVoidInflightDryRun(body)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, preview.DryRun)
	mockClient.AssertExpectations(t)
}

func TestRefundTransactionRequest_JSONIncludesDescriptionAndMeta(t *testing.T) {
	body := blnkgo.RefundTransactionRequest{
		SkipQueue:   true,
		Description: "reversed",
		MetaData:    blnkgo.MetaData{"reason": "customer"},
	}

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(payload, &decoded))
	assert.Equal(t, true, decoded["skip_queue"])
	assert.Equal(t, "reversed", decoded["description"])
	meta, ok := decoded["meta_data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "customer", meta["reason"])
}
