package blnkgo_test

import (
	"errors"
	"net/http"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func validBulkVoidInflightRequest() blnkgo.BulkVoidInflightRequest {
	return blnkgo.BulkVoidInflightRequest{
		TransactionIDs: []string{
			"txn_11111111-1111-4111-8111-111111111111",
			"txn_22222222-2222-4222-8222-222222222222",
			"txn_33333333-3333-4333-8333-333333333333",
		},
	}
}

func TestTransactionService_BulkVoidInflight_Success(t *testing.T) {
	mockClient, svc := setupTransactionService()
	body := validBulkVoidInflightRequest()

	expectedResponse := &blnkgo.BulkVoidInflightResponse{
		Succeeded: 2,
		Failed:    1,
		Results: []blnkgo.BulkVoidInflightResult{
			{TransactionID: "txn_11111111-1111-4111-8111-111111111111", Status: "succeeded"},
			{TransactionID: "txn_22222222-2222-4222-8222-222222222222", Status: "succeeded"},
			{
				TransactionID: "txn_33333333-3333-4333-8333-333333333333",
				Status:        "failed",
				Code:          "ALREADY_COMMITTED",
				Message:       "cannot void. Transaction already committed",
			},
		},
	}

	mockClient.On("NewRequest", "transactions/inflight/bulk/void", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		response := args.Get(1).(*blnkgo.BulkVoidInflightResponse)
		*response = *expectedResponse
	})

	result, resp, err := svc.BulkVoidInflight(body)

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_BulkVoidInflight_EmptyTransactionIDs(t *testing.T) {
	mockClient, svc := setupTransactionService()

	result, resp, err := svc.BulkVoidInflight(blnkgo.BulkVoidInflightRequest{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction_ids array cannot be empty")
	assert.Nil(t, result)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
	mockClient.AssertNotCalled(t, "CallWithRetry")
}

func TestTransactionService_BulkVoidInflight_TooManyTransactionIDs(t *testing.T) {
	mockClient, svc := setupTransactionService()
	ids := make([]string, blnkgo.MaxBulkInflightItems+1)
	for i := range ids {
		ids[i] = "txn_test"
	}

	result, resp, err := svc.BulkVoidInflight(blnkgo.BulkVoidInflightRequest{TransactionIDs: ids})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too many transaction_ids")
	assert.Nil(t, result)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
	mockClient.AssertNotCalled(t, "CallWithRetry")
}

func TestTransactionService_BulkVoidInflight_EmptyTransactionID(t *testing.T) {
	mockClient, svc := setupTransactionService()
	body := blnkgo.BulkVoidInflightRequest{
		TransactionIDs: []string{
			"txn_valid",
			"",
		},
	}

	result, resp, err := svc.BulkVoidInflight(body)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction_id is required at index 1")
	assert.Nil(t, result)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
	mockClient.AssertNotCalled(t, "CallWithRetry")
}

func TestTransactionService_BulkVoidInflight_NewRequestError(t *testing.T) {
	mockClient, svc := setupTransactionService()
	body := validBulkVoidInflightRequest()

	mockClient.On("NewRequest", "transactions/inflight/bulk/void", http.MethodPost, body).Return(nil, errors.New("failed to create request"))

	result, resp, err := svc.BulkVoidInflight(body)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create request")
	assert.Nil(t, result)
	assert.Nil(t, resp)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_BulkVoidInflight_ServerError(t *testing.T) {
	mockClient, svc := setupTransactionService()
	body := validBulkVoidInflightRequest()

	mockClient.On("NewRequest", "transactions/inflight/bulk/void", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusBadRequest,
	}, errors.New("bad request"))

	result, resp, err := svc.BulkVoidInflight(body)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_BulkVoidInflight_WithSkipQueue(t *testing.T) {
	mockClient, svc := setupTransactionService()
	body := blnkgo.BulkVoidInflightRequest{
		SkipQueue:      true,
		TransactionIDs: []string{"txn_11111111-1111-4111-8111-111111111111"},
	}

	mockClient.On(
		"NewRequest",
		"transactions/inflight/bulk/void",
		http.MethodPost,
		mock.MatchedBy(func(req interface{}) bool {
			bulk, ok := req.(blnkgo.BulkVoidInflightRequest)
			return ok && bulk.SkipQueue && len(bulk.TransactionIDs) == 1
		}),
	).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		response := args.Get(1).(*blnkgo.BulkVoidInflightResponse)
		*response = blnkgo.BulkVoidInflightResponse{
			Succeeded: 1,
			Results: []blnkgo.BulkVoidInflightResult{
				{TransactionID: "txn_11111111-1111-4111-8111-111111111111", Status: "succeeded"},
			},
		}
	})

	result, resp, err := svc.BulkVoidInflight(body)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, result.Succeeded)
	assert.Equal(t, "succeeded", result.Results[0].Status)
	mockClient.AssertExpectations(t)
}
