package blnkgo_test

import (
	"errors"
	"net/http"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupHooksService() (*MockClient, *blnkgo.HooksService) {
	mockClient := &MockClient{}
	svc := blnkgo.NewHooksService(mockClient)
	return mockClient, svc
}

func TestHooksService_Create_Success(t *testing.T) {
	mockClient, svc := setupHooksService()

	body := blnkgo.CreateHookRequest{
		Name:       "Pre-transaction validation",
		URL:        "https://api.example.com/validate",
		Type:       blnkgo.HookTypePreTransaction,
		Active:     true,
		Timeout:    30,
		RetryCount: 3,
	}

	mockClient.On("NewRequest", "hooks", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusCreated}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.HookResponse)
		*resp = blnkgo.HookResponse{
			ID:          "hk_test_123",
			Name:        body.Name,
			URL:         body.URL,
			Type:        body.Type,
			Active:      body.Active,
			Timeout:     body.Timeout,
			RetryCount:  body.RetryCount,
			CreatedAt:   "2024-11-26T08:36:36.238244338Z",
			LastRun:     "0001-01-01T00:00:00Z",
			LastSuccess: false,
		}
	})

	hook, httpResp, err := svc.Create(body)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusCreated, httpResp.StatusCode)
	assert.Equal(t, "hk_test_123", hook.ID)
	assert.Equal(t, body.Name, hook.Name)
	assert.Equal(t, body.URL, hook.URL)
	mockClient.AssertExpectations(t)
}

func TestHooksService_Create_ValidationError(t *testing.T) {
	mockClient, svc := setupHooksService()

	_, _, err := svc.Create(blnkgo.CreateHookRequest{
		Name:       "missing-url",
		Type:       blnkgo.HookTypePreTransaction,
		Active:     true,
		Timeout:    30,
		RetryCount: 3,
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "url is required")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestHooksService_Create_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupHooksService()

	body := blnkgo.CreateHookRequest{
		Name:       "Pre-transaction validation",
		URL:        "https://api.example.com/validate",
		Type:       blnkgo.HookTypePreTransaction,
		Active:     true,
		Timeout:    30,
		RetryCount: 3,
	}

	mockClient.On("NewRequest", "hooks", http.MethodPost, body).Return(nil, errors.New("failed to create request"))

	hook, httpResp, err := svc.Create(body)

	assert.Error(t, err)
	assert.Nil(t, hook)
	assert.Nil(t, httpResp)
	mockClient.AssertExpectations(t)
}
