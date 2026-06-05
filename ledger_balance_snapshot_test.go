package blnkgo_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLedgerBalanceService_CreateSnapshot_Success(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	body := blnkgo.CreateBalanceSnapshotRequest{BatchSize: 500}

	expectedResponse := &blnkgo.CreateBalanceSnapshotResponse{
		Message: "Snapshotting in progress. should be completed shortly",
	}

	mockClient.On("NewRequest", "balances-snapshots?batch_size=500", http.MethodPost, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		response := args.Get(1).(*blnkgo.CreateBalanceSnapshotResponse)
		*response = *expectedResponse
	})

	response, resp, err := svc.CreateSnapshot(body)

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_CreateSnapshot_DefaultBatchSize(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	body := blnkgo.CreateBalanceSnapshotRequest{}

	mockClient.On("NewRequest", "balances-snapshots", http.MethodPost, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil)

	_, _, err := svc.CreateSnapshot(body)

	assert.NoError(t, err)
	mockClient.AssertCalled(t, "NewRequest", "balances-snapshots", http.MethodPost, nil)
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_CreateSnapshot_CorrectEndpoint(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	body := blnkgo.CreateBalanceSnapshotRequest{BatchSize: 250}

	mockClient.On("NewRequest", "balances-snapshots?batch_size=250", http.MethodPost, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil)

	_, _, _ = svc.CreateSnapshot(body)

	mockClient.AssertCalled(t, "NewRequest", "balances-snapshots?batch_size=250", http.MethodPost, nil)
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_CreateSnapshot_InvalidBatchSize(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	body := blnkgo.CreateBalanceSnapshotRequest{BatchSize: -1}

	response, resp, err := svc.CreateSnapshot(body)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "batch_size must be positive")
	assert.Nil(t, response)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
	mockClient.AssertNotCalled(t, "CallWithRetry")
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_CreateSnapshot_NewRequestError(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	body := blnkgo.CreateBalanceSnapshotRequest{}

	mockClient.On("NewRequest", "balances-snapshots", http.MethodPost, nil).Return(nil, fmt.Errorf("request error"))

	response, resp, err := svc.CreateSnapshot(body)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "CallWithRetry")
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_CreateSnapshot_ServerError(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	body := blnkgo.CreateBalanceSnapshotRequest{}
	expectedResp := &http.Response{StatusCode: http.StatusInternalServerError}

	mockClient.On("NewRequest", "balances-snapshots", http.MethodPost, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(expectedResp, fmt.Errorf("server error"))

	response, resp, err := svc.CreateSnapshot(body)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestCreateBalanceSnapshotResponse_UnmarshalJSON(t *testing.T) {
	// Core v0.14.3 TakeBalanceSnapshots response: {"message":"Snapshotting in progress. should be completed shortly"}
	payload := []byte(`{"message":"Snapshotting in progress. should be completed shortly"}`)

	var response blnkgo.CreateBalanceSnapshotResponse
	err := json.Unmarshal(payload, &response)

	assert.NoError(t, err)
	assert.Equal(t, "Snapshotting in progress. should be completed shortly", response.Message)

	// Round-trip confirms struct tags match the API field name.
	encoded, err := json.Marshal(response)
	assert.NoError(t, err)
	assert.JSONEq(t, string(payload), string(encoded))
}
