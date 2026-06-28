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

func TestHooksService_Update_Success(t *testing.T) {
	mockClient, svc := setupHooksService()

	hookID := "hk_test_123"
	body := blnkgo.UpdateHookRequest{
		Name:       "Pre-transaction validation (updated)",
		URL:        "https://api.example.com/validate-v2",
		Type:       blnkgo.HookTypePreTransaction,
		Active:     false,
		Timeout:    45,
		RetryCount: 5,
	}

	mockClient.On("NewRequest", "hooks/"+hookID, http.MethodPut, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.HookResponse)
		*resp = blnkgo.HookResponse{
			ID:          hookID,
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

	hook, httpResp, err := svc.Update(hookID, body)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, hookID, hook.ID)
	assert.Equal(t, body.Name, hook.Name)
	assert.False(t, hook.Active)
	mockClient.AssertExpectations(t)
}

func TestHooksService_Update_ValidationError(t *testing.T) {
	mockClient, svc := setupHooksService()

	_, _, err := svc.Update("", blnkgo.UpdateHookRequest{
		Name:       "updated",
		URL:        "https://api.example.com/validate",
		Type:       blnkgo.HookTypePreTransaction,
		Active:     true,
		Timeout:    30,
		RetryCount: 3,
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hook id is required")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestHooksService_Update_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupHooksService()

	hookID := "hk_test_123"
	body := blnkgo.UpdateHookRequest{
		Name:       "Pre-transaction validation (updated)",
		URL:        "https://api.example.com/validate-v2",
		Type:       blnkgo.HookTypePreTransaction,
		Active:     false,
		Timeout:    45,
		RetryCount: 5,
	}

	mockClient.On("NewRequest", "hooks/"+hookID, http.MethodPut, body).Return(nil, errors.New("failed to create request"))

	hook, httpResp, err := svc.Update(hookID, body)

	assert.Error(t, err)
	assert.Nil(t, hook)
	assert.Nil(t, httpResp)
	mockClient.AssertExpectations(t)
}

func TestHooksService_Get_Success(t *testing.T) {
	mockClient, svc := setupHooksService()

	hookID := "hk_test_123"
	mockClient.On("NewRequest", "hooks/"+hookID, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.HookResponse)
		*resp = blnkgo.HookResponse{
			ID:          hookID,
			Name:        "Pre-transaction validation",
			URL:         "https://api.example.com/validate",
			Type:        blnkgo.HookTypePreTransaction,
			Active:      true,
			Timeout:     30,
			RetryCount:  3,
			CreatedAt:   "2024-11-26T08:36:36.238244338Z",
			LastRun:     "0001-01-01T00:00:00Z",
			LastSuccess: false,
		}
	})

	hook, httpResp, err := svc.Get(hookID)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, hookID, hook.ID)
	mockClient.AssertExpectations(t)
}

func TestHooksService_Get_ValidationError(t *testing.T) {
	mockClient, svc := setupHooksService()

	_, _, err := svc.Get("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hook id is required")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestHooksService_Get_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupHooksService()

	hookID := "hk_test_123"
	mockClient.On("NewRequest", "hooks/"+hookID, http.MethodGet, nil).Return(nil, errors.New("failed to create request"))

	hook, httpResp, err := svc.Get(hookID)

	assert.Error(t, err)
	assert.Nil(t, hook)
	assert.Nil(t, httpResp)
	mockClient.AssertExpectations(t)
}

func TestHooksService_List_Success(t *testing.T) {
	mockClient, svc := setupHooksService()

	opts := &blnkgo.ListHooksOptions{Type: blnkgo.HookTypePreTransaction}
	mockClient.On("NewRequest", "hooks", http.MethodGet, opts).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		hooks := args.Get(1).(*[]blnkgo.HookResponse)
		*hooks = []blnkgo.HookResponse{
			{
				ID:          "hook_test_123",
				Name:        "Pre-transaction validation",
				URL:         "https://api.example.com/validate",
				Type:        blnkgo.HookTypePreTransaction,
				Active:      true,
				Timeout:     30,
				RetryCount:  3,
				CreatedAt:   "2024-11-26T08:36:36.238244338Z",
				LastRun:     "0001-01-01T00:00:00Z",
				LastSuccess: false,
			},
		}
	})

	hooks, httpResp, err := svc.List(opts)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Len(t, hooks, 1)
	assert.Equal(t, "hook_test_123", hooks[0].ID)
	mockClient.AssertExpectations(t)
}

func TestHooksService_List_WithoutOptions(t *testing.T) {
	mockClient, svc := setupHooksService()

	mockClient.On("NewRequest", "hooks", http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		hooks := args.Get(1).(*[]blnkgo.HookResponse)
		*hooks = []blnkgo.HookResponse{}
	})

	hooks, httpResp, err := svc.List(nil)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Empty(t, hooks)
	mockClient.AssertExpectations(t)
}

func TestHooksService_List_ValidationError(t *testing.T) {
	mockClient, svc := setupHooksService()

	_, _, err := svc.List(&blnkgo.ListHooksOptions{Type: "INVALID"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "type must be PRE_TRANSACTION or POST_TRANSACTION")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestHooksService_List_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupHooksService()

	opts := &blnkgo.ListHooksOptions{Type: blnkgo.HookTypePreTransaction}
	mockClient.On("NewRequest", "hooks", http.MethodGet, opts).Return(nil, errors.New("failed to create request"))

	hooks, httpResp, err := svc.List(opts)

	assert.Error(t, err)
	assert.Nil(t, hooks)
	assert.Nil(t, httpResp)
	mockClient.AssertExpectations(t)
}

func TestHooksService_Delete_Success(t *testing.T) {
	mockClient, svc := setupHooksService()

	hookID := "hook_test_123"
	mockClient.On("NewRequest", "hooks/"+hookID, http.MethodDelete, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.DeleteHookResponse)
		*resp = blnkgo.DeleteHookResponse{Message: "hook deleted successfully"}
	})

	deleted, httpResp, err := svc.Delete(hookID)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, "hook deleted successfully", deleted.Message)
	mockClient.AssertExpectations(t)
}

func TestHooksService_Delete_ValidationError(t *testing.T) {
	mockClient, svc := setupHooksService()

	_, _, err := svc.Delete("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hook id is required")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestHooksService_Delete_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupHooksService()

	hookID := "hook_test_123"
	mockClient.On("NewRequest", "hooks/"+hookID, http.MethodDelete, nil).Return(nil, errors.New("failed to create request"))

	deleted, httpResp, err := svc.Delete(hookID)

	assert.Error(t, err)
	assert.Nil(t, deleted)
	assert.Nil(t, httpResp)
	mockClient.AssertExpectations(t)
}
