package blnkgo_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
