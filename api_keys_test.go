package blnkgo_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupApiKeysService() (*MockClient, *blnkgo.ApiKeysService) {
	mockClient := &MockClient{}
	svc := blnkgo.NewApiKeysService(mockClient)
	return mockClient, svc
}

func TestApiKeysService_Create_Success(t *testing.T) {
	mockClient, svc := setupApiKeysService()

	expiresAt := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	body := blnkgo.CreateApiKeyRequest{
		Name:      "integration-key",
		Owner:     "owner_test",
		Scopes:    []string{"ledgers:read"},
		ExpiresAt: expiresAt,
	}

	mockClient.On("NewRequest", "api-keys", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusCreated}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.ApiKeyResponse)
		*resp = blnkgo.ApiKeyResponse{
			ApiKeyID:  "api_key_abc123",
			Key:       "secret-key-value",
			Name:      body.Name,
			OwnerID:   body.Owner,
			Scopes:    body.Scopes,
			ExpiresAt: expiresAt.Format(time.RFC3339),
			CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
			IsRevoked: false,
		}
	})

	apiKeyResp, httpResp, err := svc.Create(body)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusCreated, httpResp.StatusCode)
	assert.Equal(t, "api_key_abc123", apiKeyResp.ApiKeyID)
	assert.Equal(t, "secret-key-value", apiKeyResp.Key)
	assert.Equal(t, body.Owner, apiKeyResp.OwnerID)
	mockClient.AssertExpectations(t)
}

func TestApiKeysService_Create_ValidationError(t *testing.T) {
	mockClient, svc := setupApiKeysService()

	_, _, err := svc.Create(blnkgo.CreateApiKeyRequest{
		Name:  "missing-fields",
		Owner: "owner_test",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one scope must be specified")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestApiKeysService_Create_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupApiKeysService()

	body := blnkgo.CreateApiKeyRequest{
		Name:      "test-key",
		Owner:     "owner_test",
		Scopes:    []string{"ledgers:read"},
		ExpiresAt: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	mockClient.On("NewRequest", "api-keys", http.MethodPost, body).Return(nil, errors.New("failed to create request"))

	apiKeyResp, httpResp, err := svc.Create(body)

	assert.Error(t, err)
	assert.Nil(t, apiKeyResp)
	assert.Nil(t, httpResp)
	mockClient.AssertExpectations(t)
}

func TestApiKeysService_List_Success(t *testing.T) {
	mockClient, svc := setupApiKeysService()

	opts := &blnkgo.ListApiKeysOptions{Owner: "owner_test"}
	mockClient.On("NewRequest", "api-keys", http.MethodGet, opts).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		keys := args.Get(1).(*[]blnkgo.ApiKeyResponse)
		*keys = []blnkgo.ApiKeyResponse{
			{
				ApiKeyID:  "api_key_abc123",
				Name:      "read-only",
				OwnerID:   "owner_test",
				Scopes:    []string{"ledgers:read"},
				IsRevoked: false,
			},
		}
	})

	keys, httpResp, err := svc.List(opts)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.Len(t, keys, 1)
	assert.Equal(t, "api_key_abc123", keys[0].ApiKeyID)
	assert.Equal(t, "owner_test", keys[0].OwnerID)
	mockClient.AssertExpectations(t)
}

func TestApiKeysService_List_WithoutOptions(t *testing.T) {
	mockClient, svc := setupApiKeysService()

	mockClient.On("NewRequest", "api-keys", http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		keys := args.Get(1).(*[]blnkgo.ApiKeyResponse)
		*keys = []blnkgo.ApiKeyResponse{}
	})

	keys, httpResp, err := svc.List(nil)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Empty(t, keys)
	mockClient.AssertExpectations(t)
}

func TestApiKeysService_List_ValidationError(t *testing.T) {
	mockClient, svc := setupApiKeysService()

	_, _, err := svc.List(&blnkgo.ListApiKeysOptions{Owner: "   "})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "owner must be a non-empty string")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestApiKeysService_List_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupApiKeysService()

	opts := &blnkgo.ListApiKeysOptions{Owner: "owner_test"}
	mockClient.On("NewRequest", "api-keys", http.MethodGet, opts).Return(nil, errors.New("failed to create request"))

	keys, httpResp, err := svc.List(opts)

	assert.Error(t, err)
	assert.Nil(t, keys)
	assert.Nil(t, httpResp)
	mockClient.AssertExpectations(t)
}

func TestApiKeysService_Delete_Success(t *testing.T) {
	mockClient, svc := setupApiKeysService()

	apiKeyID := "api_key_abc123"
	opts := &blnkgo.DeleteApiKeysOptions{Owner: "owner_test"}
	mockClient.On("NewRequest", "api-keys/"+apiKeyID+"?owner=owner_test", http.MethodDelete, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusNoContent}, nil)

	httpResp, err := svc.Delete(apiKeyID, opts)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusNoContent, httpResp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestApiKeysService_Delete_WithoutOptions(t *testing.T) {
	mockClient, svc := setupApiKeysService()

	apiKeyID := "api_key_abc123"
	mockClient.On("NewRequest", "api-keys/"+apiKeyID, http.MethodDelete, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusNoContent}, nil)

	httpResp, err := svc.Delete(apiKeyID, nil)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, httpResp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestApiKeysService_Delete_ValidationError(t *testing.T) {
	mockClient, svc := setupApiKeysService()

	httpResp, err := svc.Delete("", nil)

	assert.Error(t, err)
	assert.Nil(t, httpResp)
	assert.Contains(t, err.Error(), "api key id is required")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestApiKeysService_Delete_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupApiKeysService()

	apiKeyID := "api_key_abc123"
	mockClient.On("NewRequest", "api-keys/"+apiKeyID, http.MethodDelete, nil).Return(nil, errors.New("failed to create request"))

	httpResp, err := svc.Delete(apiKeyID, nil)

	assert.Error(t, err)
	assert.Nil(t, httpResp)
	mockClient.AssertExpectations(t)
}
