package blnkgo_test

import (
	"errors"
	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"testing"
)

// Helper function to setup mock client and service
func setupMetdataService() (*MockClient, *blnkgo.MetadataService) {
	mockClient := &MockClient{}
	svc := blnkgo.NewMetadataService(mockClient)
	return mockClient, svc
}

func TestTransactionService_UpdateMetadata(t *testing.T) {
	tests := []struct {
		name        string
		entityID    string
		metadata    map[string]interface{}
		expectError bool
		errorMsg    string
		statusCode  int
		setupMocks  func(*MockClient)
	}{
		{
			name:     "successful metadata update",
			entityID: "entity-123",
			metadata: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
				"key3": true,
			},
			expectError: false,
			statusCode:  http.StatusOK,
			setupMocks: func(m *MockClient) {
				expectedResponse := &blnkgo.UpdateMetaDataRequest{
					MetaData: map[string]interface{}{
						"key1": "value1",
						"key2": 42,
						"key3": true,
					},
				}

				m.On("NewRequest", "entity-123/metadata", http.MethodPost,
					mock.MatchedBy(func(req blnkgo.UpdateMetaDataRequest) bool {
						return assert.ObjectsAreEqual(req.MetaData, expectedResponse.MetaData)
					})).Return(&http.Request{}, nil)

				m.On("CallWithRetry", mock.Anything, mock.Anything).
					Return(&http.Response{StatusCode: http.StatusOK}, nil).
					Run(func(args mock.Arguments) {
						response := args.Get(1).(*blnkgo.Metadata)
						*response = blnkgo.Metadata(*expectedResponse)
					})
			},
		},
		{
			name:        "empty entity ID",
			entityID:    "",
			metadata:    map[string]interface{}{"key": "value"},
			expectError: true,
			errorMsg:    "entity ID is required",
			setupMocks:  func(m *MockClient) {},
		},
		{
			name:        "request creation failure",
			entityID:    "entity-123",
			metadata:    map[string]interface{}{"key": "value"},
			expectError: true,
			errorMsg:    "failed to create request",
			setupMocks: func(m *MockClient) {
				m.On("NewRequest", "entity-123/metadata", http.MethodPost, mock.Anything).
					Return(nil, errors.New("failed to create request"))
			},
		},
		{
			name:        "server error",
			entityID:    "entity-123",
			metadata:    map[string]interface{}{"key": "value"},
			expectError: true,
			errorMsg:    "server error",
			statusCode:  http.StatusInternalServerError,
			setupMocks: func(m *MockClient) {
				m.On("NewRequest", "entity-123/metadata", http.MethodPost, mock.Anything).
					Return(&http.Request{}, nil)
				m.On("CallWithRetry", mock.Anything, mock.Anything).
					Return(&http.Response{StatusCode: http.StatusInternalServerError},
						errors.New("server error"))
			},
		},
		{
			name:        "entity not found",
			entityID:    "nonexistent-123",
			metadata:    map[string]interface{}{"key": "value"},
			expectError: true,
			errorMsg:    "entity not found",
			statusCode:  http.StatusNotFound,
			setupMocks: func(m *MockClient) {
				m.On("NewRequest", "nonexistent-123/metadata", http.MethodPost, mock.Anything).
					Return(&http.Request{}, nil)
				m.On("CallWithRetry", mock.Anything, mock.Anything).
					Return(&http.Response{StatusCode: http.StatusNotFound},
						errors.New("entity not found"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient, svc := setupMetdataService()
			tt.setupMocks(mockClient)

			response, resp, err := svc.UpdateMetadata(tt.entityID, blnkgo.UpdateMetaDataRequest{MetaData: tt.metadata})

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				if tt.entityID == "" {
					assert.Nil(t, resp)
					mockClient.AssertNotCalled(t, "NewRequest")
					mockClient.AssertNotCalled(t, "CallWithRetry")
				} else if tt.name == "request creation failure" {
					assert.Nil(t, response)
					assert.Nil(t, resp)
					mockClient.AssertNotCalled(t, "CallWithRetry")
				} else {
					assert.Equal(t, tt.statusCode, resp.StatusCode)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.statusCode, resp.StatusCode)
				assert.Equal(t, tt.metadata, response.MetaData)
			}
			mockClient.AssertExpectations(t)
		})
	}
}
