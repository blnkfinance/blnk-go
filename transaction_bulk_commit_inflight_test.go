package blnkgo_test

import (
	"errors"
	"math/big"
	"net/http"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func validBulkCommitInflightRequest() blnkgo.BulkCommitInflightRequest {
	return blnkgo.BulkCommitInflightRequest{
		Transactions: []blnkgo.BulkCommitInflightItem{
			{TransactionID: "txn_11111111-1111-4111-8111-111111111111"},
			{TransactionID: "txn_22222222-2222-4222-8222-222222222222", Amount: 40},
			{
				TransactionID: "txn_33333333-3333-4333-8333-333333333333",
				PreciseAmount: big.NewInt(125034),
			},
		},
	}
}

func TestTransactionService_BulkCommitInflight_Success(t *testing.T) {
	mockClient, svc := setupTransactionService()
	body := validBulkCommitInflightRequest()

	expectedResponse := &blnkgo.BulkCommitInflightResponse{
		Succeeded: 2,
		Failed:    1,
		Results: []blnkgo.BulkCommitInflightResult{
			{TransactionID: "txn_11111111-1111-4111-8111-111111111111", Status: "succeeded"},
			{TransactionID: "txn_22222222-2222-4222-8222-222222222222", Status: "succeeded"},
			{
				TransactionID: "txn_33333333-3333-4333-8333-333333333333",
				Status:        "failed",
				Code:          "INVALID_AMOUNT",
				Message:       "cannot commit more than inflight amount",
			},
		},
	}

	mockClient.On("NewRequest", "transactions/inflight/bulk/commit", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		response := args.Get(1).(*blnkgo.BulkCommitInflightResponse)
		*response = *expectedResponse
	})

	result, resp, err := svc.BulkCommitInflight(body)

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_BulkCommitInflight_EmptyTransactions(t *testing.T) {
	mockClient, svc := setupTransactionService()

	result, resp, err := svc.BulkCommitInflight(blnkgo.BulkCommitInflightRequest{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transactions array cannot be empty")
	assert.Nil(t, result)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
	mockClient.AssertNotCalled(t, "CallWithRetry")
}

func TestTransactionService_BulkCommitInflight_TooManyTransactions(t *testing.T) {
	mockClient, svc := setupTransactionService()
	items := make([]blnkgo.BulkCommitInflightItem, blnkgo.MaxBulkInflightItems+1)
	for i := range items {
		items[i] = blnkgo.BulkCommitInflightItem{TransactionID: "txn_test"}
	}

	result, resp, err := svc.BulkCommitInflight(blnkgo.BulkCommitInflightRequest{Transactions: items})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too many transactions")
	assert.Nil(t, result)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
	mockClient.AssertNotCalled(t, "CallWithRetry")
}

func TestTransactionService_BulkCommitInflight_EmptyTransactionID(t *testing.T) {
	mockClient, svc := setupTransactionService()
	body := blnkgo.BulkCommitInflightRequest{
		Transactions: []blnkgo.BulkCommitInflightItem{
			{TransactionID: "txn_valid"},
			{TransactionID: ""},
		},
	}

	result, resp, err := svc.BulkCommitInflight(body)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction_id is required at index 1")
	assert.Nil(t, result)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
	mockClient.AssertNotCalled(t, "CallWithRetry")
}

func TestTransactionService_BulkCommitInflight_NewRequestError(t *testing.T) {
	mockClient, svc := setupTransactionService()
	body := validBulkCommitInflightRequest()

	mockClient.On("NewRequest", "transactions/inflight/bulk/commit", http.MethodPost, body).Return(nil, errors.New("failed to create request"))

	result, resp, err := svc.BulkCommitInflight(body)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create request")
	assert.Nil(t, result)
	assert.Nil(t, resp)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_BulkCommitInflight_ServerError(t *testing.T) {
	mockClient, svc := setupTransactionService()
	body := validBulkCommitInflightRequest()

	mockClient.On("NewRequest", "transactions/inflight/bulk/commit", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusBadRequest,
	}, errors.New("bad request"))

	result, resp, err := svc.BulkCommitInflight(body)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_BulkCommitInflight_WithSkipQueue(t *testing.T) {
	mockClient, svc := setupTransactionService()
	body := blnkgo.BulkCommitInflightRequest{
		SkipQueue: true,
		Transactions: []blnkgo.BulkCommitInflightItem{
			{TransactionID: "txn_11111111-1111-4111-8111-111111111111"},
		},
	}

	mockClient.On(
		"NewRequest",
		"transactions/inflight/bulk/commit",
		http.MethodPost,
		mock.MatchedBy(func(req interface{}) bool {
			bulk, ok := req.(blnkgo.BulkCommitInflightRequest)
			return ok && bulk.SkipQueue && len(bulk.Transactions) == 1
		}),
	).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		response := args.Get(1).(*blnkgo.BulkCommitInflightResponse)
		*response = blnkgo.BulkCommitInflightResponse{
			Succeeded: 1,
			Results: []blnkgo.BulkCommitInflightResult{
				{TransactionID: "txn_11111111-1111-4111-8111-111111111111", Status: "succeeded"},
			},
		}
	})

	result, resp, err := svc.BulkCommitInflight(body)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, result.Succeeded)
	assert.Equal(t, "succeeded", result.Results[0].Status)
	mockClient.AssertExpectations(t)
}
