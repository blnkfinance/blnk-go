package blnkgo_test

import (
	"encoding/json"
	"net/http"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransactionService_RecoverQueue_Success(t *testing.T) {
	mockClient, svc := setupTransactionService()

	expectedResponse := &blnkgo.RecoverQueueResponse{
		Recovered: 2,
		Threshold: "2m0s",
	}

	mockClient.On("NewRequest", "transactions/recover", http.MethodPost, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		response := args.Get(1).(*blnkgo.RecoverQueueResponse)
		*response = *expectedResponse
	})

	result, resp, err := svc.RecoverQueue(blnkgo.RecoverQueueRequest{})

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_RecoverQueue_WithThreshold(t *testing.T) {
	mockClient, svc := setupTransactionService()

	expectedResponse := &blnkgo.RecoverQueueResponse{
		Recovered: 0,
		Threshold: "24h0m0s",
	}

	mockClient.On("NewRequest", "transactions/recover?threshold=5m", http.MethodPost, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		response := args.Get(1).(*blnkgo.RecoverQueueResponse)
		*response = *expectedResponse
	})

	result, resp, err := svc.RecoverQueue(blnkgo.RecoverQueueRequest{Threshold: "5m"})

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertCalled(t, "NewRequest", "transactions/recover?threshold=5m", http.MethodPost, nil)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_RecoverQueue_InvalidThreshold(t *testing.T) {
	mockClient, svc := setupTransactionService()

	result, resp, err := svc.RecoverQueue(blnkgo.RecoverQueueRequest{Threshold: "bogus"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "threshold must be a valid duration string")
	assert.Nil(t, result)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
	mockClient.AssertNotCalled(t, "CallWithRetry")
}

func TestTransactionService_RecoverQueue_NewRequestError(t *testing.T) {
	mockClient, svc := setupTransactionService()

	mockClient.On("NewRequest", "transactions/recover", http.MethodPost, nil).Return((*http.Request)(nil), assert.AnError)

	result, resp, err := svc.RecoverQueue(blnkgo.RecoverQueueRequest{})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Nil(t, resp)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_RecoverQueue_ServerError(t *testing.T) {
	mockClient, svc := setupTransactionService()

	mockClient.On("NewRequest", "transactions/recover", http.MethodPost, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return((*http.Response)(nil), assert.AnError)

	result, resp, err := svc.RecoverQueue(blnkgo.RecoverQueueRequest{})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Nil(t, resp)
	mockClient.AssertExpectations(t)
}

func TestRecoverQueueResponse_UnmarshalJSON(t *testing.T) {
	payload := []byte(`{"recovered":3,"threshold":"5m0s"}`)

	var response blnkgo.RecoverQueueResponse
	err := json.Unmarshal(payload, &response)

	assert.NoError(t, err)
	assert.Equal(t, 3, response.Recovered)
	assert.Equal(t, "5m0s", response.Threshold)
}
