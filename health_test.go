package blnkgo_test

import (
	"net/http"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupHealthService() (*MockClient, *blnkgo.HealthService) {
	mockClient := &MockClient{}
	svc := blnkgo.NewHealthService(mockClient)
	return mockClient, svc
}

func TestHealthService_Check_Success(t *testing.T) {
	mockClient, svc := setupHealthService()

	mockClient.On("NewRequest", "health", http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.HealthResponse)
		*resp = blnkgo.HealthResponse{Status: "UP"}
	})

	healthResp, httpResp, err := svc.Check()

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, "UP", healthResp.Status)
	mockClient.AssertExpectations(t)
}

func TestHealthService_Check_ServiceUnavailable(t *testing.T) {
	mockClient, svc := setupHealthService()

	mockClient.On("NewRequest", "health", http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusServiceUnavailable}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.HealthResponse)
		*resp = blnkgo.HealthResponse{Status: "DOWN", Reason: "database ping failed"}
	})

	healthResp, httpResp, err := svc.Check()

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusServiceUnavailable, httpResp.StatusCode)
	assert.Equal(t, "DOWN", healthResp.Status)
	assert.Equal(t, "database ping failed", healthResp.Reason)
	mockClient.AssertExpectations(t)
}

func TestHealthService_Check_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupHealthService()

	mockClient.On("NewRequest", "health", http.MethodGet, nil).Return(nil, assert.AnError)

	healthResp, httpResp, err := svc.Check()

	assert.Error(t, err)
	assert.Nil(t, healthResp)
	assert.Nil(t, httpResp)
	mockClient.AssertExpectations(t)
}
