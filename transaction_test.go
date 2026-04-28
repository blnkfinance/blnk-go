package blnkgo_test

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

type Tests[T any] struct {
	name        string
	body        T
	expectError bool
	errorMsg    string
}

// NewRequest is a mock method that simulates creating a new HTTP request.
func (m *MockClient) NewRequest(endpoint string, method string, body interface{}) (*http.Request, error) {
	args := m.Called(endpoint, method, body)
	if req, ok := args.Get(0).(*http.Request); ok || args.Get(0) == nil {
		return req, args.Error(1)
	}

	return nil, args.Error(1)
}

// CallWithRetry is a mock method that simulates making an HTTP call with retry logic.
func (m *MockClient) CallWithRetry(req *http.Request, v interface{}) (*http.Response, error) {
	args := m.Called(req, v)
	if resp, ok := args.Get(0).(*http.Response); ok || args.Get(0) == nil {
		return resp, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockClient) NewFileUploadRequest(endpoint string, fileParam string, file interface{}, fileName string, fields map[string]string) (*http.Request, error) {
	args := m.Called(endpoint, fileParam, file, fileName, fields)
	if req, ok := args.Get(0).(*http.Request); ok || args.Get(0) == nil {
		return req, args.Error(1)
	}
	return nil, args.Error(1)
}

// Helper function to setup mock client and service
func setupTransactionService() (*MockClient, *blnkgo.TransactionService) {
	mockClient := &MockClient{}
	svc := blnkgo.NewTransactionService(mockClient)
	return mockClient, svc
}
func TestUpdateTransaction(t *testing.T) {
	// Setup mock client and service
	tests := []Tests[blnkgo.UpdateStatus]{
		{
			name: "Update transaction Success",
			body: blnkgo.UpdateStatus{
				Status: blnkgo.InflightStatusCommit,
			},
			expectError: false,
			errorMsg:    "",
		},
		{
			name: "Update transaction Fail",
			body: blnkgo.UpdateStatus{
				Status: blnkgo.InflightStatusVoid,
			},
			expectError: true,
			errorMsg:    "transaction not found",
		},
		{
			name: "Valid Url Format",
			body: blnkgo.UpdateStatus{
				Status: blnkgo.InflightStatusCommit,
			},
			expectError: false,
			errorMsg:    "",
		},
		{
			name: "Invalid Url Format",
			body: blnkgo.UpdateStatus{
				Status: blnkgo.InflightStatusCommit,
			},
			expectError: true,
			errorMsg:    "invalid URL format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient, svc := setupTransactionService()

			// Setup expected response with fixed time
			fixedTime := time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC)
			effectiveDate := time.Date(2023, time.September, 15, 10, 0, 0, 0, time.UTC)
			expectedResp := &blnkgo.Transaction{
				ParentTransaction: blnkgo.ParentTransaction{
					Amount:      1000,
					Reference:   "ref-21",
					Precision:   100,
					Currency:    "USD",
					Source:      "@bank-account",
					Destination: "@World",
					MetaData: map[string]interface{}{
						"transaction_type": "deposit",
						"customer_name":    "Alice Johnson",
						"customer_id":      "alice-5786",
					},
					Description:   "Alice Funds",
					Status:        blnkgo.PryTransactionStatus(tt.body.Status),
					EffectiveDate: &effectiveDate,
				},
				TransactionID: "tx-123",
				CreatedAt:     fixedTime,
			}

			// Setup mock expectations
			mockClient.On("NewRequest", "transactions/inflight/tx-123", http.MethodPut, tt.body).Return(&http.Request{}, nil)

			// Setup mock expectations for CallWithRetry
			if tt.expectError {
				mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("%s", tt.errorMsg))
			} else {
				mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(
					&http.Response{},
					nil,
				).Run(func(args mock.Arguments) {
					transaction := args.Get(1).(*blnkgo.Transaction)
					*transaction = *expectedResp
				})
			}

			transaction, resp, err := svc.Update("tx-123", tt.body)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, transaction)
				assert.Nil(t, resp)

			} else {
				// Assert
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, expectedResp, transaction)
				mockClient.AssertExpectations(t)
			}
		})
	}
}

func TestTransactionService_Update_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupTransactionService()

	// Setup mock expectations for NewRequest to return an error
	mockClient.On("NewRequest", mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to create request"))

	// Call the Update method
	transaction, resp, err := svc.Update("tx-123", blnkgo.UpdateStatus{
		Status: blnkgo.InflightStatusCommit,
	})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create request")
	assert.Nil(t, transaction)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "CallWithRetry")
}

func TestTransactionService_Update_InvalidID(t *testing.T) {
	mockClient, svc := setupTransactionService()

	// Call the Update method with an invalid ID
	transaction, resp, err := svc.Update("", blnkgo.UpdateStatus{
		Status: blnkgo.InflightStatusCommit,
	})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transactionID is required")
	assert.Nil(t, transaction)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "CallWithRetry")
}

func TestTransactionService_Update_ServerError(t *testing.T) {
	mockClient, svc := setupTransactionService()

	body := blnkgo.UpdateStatus{
		Status: blnkgo.InflightStatusCommit,
	}
	// Setup mock expectations for NewRequest
	mockClient.On("NewRequest", "transactions/inflight/123", http.MethodPut, body).Return(&http.Request{}, nil)

	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusInternalServerError}, fmt.Errorf("internal server error"))

	transaction, resp, err := svc.Update("123", blnkgo.UpdateStatus{
		Status: blnkgo.InflightStatusCommit,
	})

	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	mockClient.AssertExpectations(t)
}

// Handle concurrent update requests for same transaction
func TestTransactionService_Update_ConcurrentRequests(t *testing.T) {
	mockClient, svc := setupTransactionService()

	body := blnkgo.UpdateStatus{
		Status: blnkgo.InflightStatusCommit,
	}

	mockClient.On("NewRequest", "transactions/inflight/123", http.MethodPut, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		transaction := args.Get(1).(*blnkgo.Transaction)
		*transaction = blnkgo.Transaction{
			ParentTransaction: blnkgo.ParentTransaction{
				Status: blnkgo.PryTransactionStatusCommit,
			},
		}
	})

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		transaction, resp, err := svc.Update("123", body)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, blnkgo.PryTransactionStatusCommit, transaction.Status)
	}()

	go func() {
		defer wg.Done()
		transaction, resp, err := svc.Update("123", body)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, blnkgo.PryTransactionStatusCommit, transaction.Status)
	}()

	wg.Wait()
	mockClient.AssertExpectations(t)
}

func TestCreateTransactionSuccess(t *testing.T) {
	mockClient, svc := setupTransactionService()
	effectiveDate := time.Date(2023, time.September, 20, 14, 30, 0, 0, time.UTC)
	body := blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      1000,
			Reference:   "ref-21",
			Precision:   100,
			Currency:    "USD",
			Source:      "@bank-account",
			Destination: "@World",
			MetaData: map[string]interface{}{
				"transaction_type": "deposit",
				"customer_name":    "Alice Johnson",
				"customer_id":      "alice-5786",
			},
			Description:   "Alice Funds",
			EffectiveDate: &effectiveDate,
		},
		Inflight: true,
	}
	fixedTime := time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC)

	mockClient.On("NewRequest", "transactions", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusCreated}, nil).Run(func(args mock.Arguments) {
		transaction := args.Get(1).(*blnkgo.Transaction)
		*transaction = blnkgo.Transaction{
			ParentTransaction: body.ParentTransaction,
			TransactionID:     "txn-123",
			CreatedAt:         fixedTime,
		}
	})

	transaction, resp, err := svc.Create(body)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "txn-123", transaction.TransactionID)
	assert.Equal(t, fixedTime, transaction.CreatedAt)

	mockClient.AssertExpectations(t)
}

func TestCreateTransactionInvalidRequest(t *testing.T) {
	mockClient, svc := setupTransactionService()
	body := blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Reference: "TEST-REF",
			Status:    blnkgo.PryTransactionStatusApplied,
		},
	}

	transaction, resp, err := svc.Create(body)
	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.Nil(t, resp)

	mockClient.AssertNotCalled(t, "NewRequest")
	mockClient.AssertNotCalled(t, "CallWithRetry")
}

func TestCreateTransactionClientError(t *testing.T) {
	mockClient, svc := setupTransactionService()
	body := blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      1000,
			Reference:   "ref-21",
			Precision:   100,
			Currency:    "USD",
			Source:      "@bank-account",
			Destination: "@World",
			Description: "",
		},
	}

	mockClient.On("NewRequest", "transactions", http.MethodPost, body).Return(nil, errors.New("failed to create request"))
	transaction, resp, err := svc.Create(body)

	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to create request")
	mockClient.AssertNotCalled(t, "CallWithRetry")
	mockClient.AssertExpectations(t)
}

func TestCreate_ServerError(t *testing.T) {
	mockClient, svc := setupTransactionService()
	body := blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      1000,
			Reference:   "ref-21",
			Precision:   100,
			Currency:    "USD",
			Source:      "@bank-account",
			Destination: "@World",
			Description: "",
		},
	}

	mockClient.On("NewRequest", "transactions", http.MethodPost, body).Return(&http.Request{}, nil)

	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusInternalServerError}, errors.New("server error"))

	transaction, resp, err := svc.Create(body)

	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, err.Error(), "server error")

	mockClient.AssertExpectations(t)
}

func TestRefundTransaction(t *testing.T) {
	mockClient, svc := setupTransactionService()
	effectiveDate := time.Date(2023, time.August, 10, 16, 45, 0, 0, time.UTC)
	body := blnkgo.Transaction{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:        1000,
			Reference:     "ref-21",
			Precision:     100,
			Currency:      "USD",
			Source:        "@bank-account",
			Destination:   "@World",
			Description:   "",
			EffectiveDate: &effectiveDate,
		},
	}
	fixedTime := time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC)

	mockClient.On("NewRequest", "refund-transaction/txn-123", http.MethodPost, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusCreated}, nil).Run(func(args mock.Arguments) {
		transaction := args.Get(1).(*blnkgo.Transaction)
		*transaction = blnkgo.Transaction{
			ParentTransaction: body.ParentTransaction,
			TransactionID:     "txn-123",
			CreatedAt:         fixedTime,
		}
	})

	transaction, resp, err := svc.Refund("txn-123")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "txn-123", transaction.TransactionID)
	assert.Equal(t, fixedTime, transaction.CreatedAt)

	mockClient.AssertExpectations(t)
}

func TestRefundTransaction_FailedRequest(t *testing.T) {
	mockClient, svc := setupTransactionService()

	mockClient.On("NewRequest", "refund-transaction/txn-123", http.MethodPost, nil).Return(nil, errors.New("failed to create request"))

	transaction, resp, err := svc.Refund("txn-123")

	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to create request")

	mockClient.AssertNotCalled(t, "CallWithRetry")
	mockClient.AssertExpectations(t)
}

func TestRefundTransaction_ClientError(t *testing.T) {
	mockClient, svc := setupTransactionService()

	mockClient.On("NewRequest", "refund-transaction/txn-123", http.MethodPost, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusBadRequest}, errors.New("client error"))

	transaction, resp, err := svc.Refund("txn-123")

	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, err.Error(), "client error")

	mockClient.AssertExpectations(t)
}

func TestTransactionService_Get(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		expectError bool
		errorMsg    string
		statusCode  int
		setupMocks  func(*MockClient)
	}{
		{
			name:        "successful get",
			id:          "tx-123",
			expectError: false,
			statusCode:  http.StatusOK,
			setupMocks: func(m *MockClient) {
				fixedTime := time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC)
				effectiveDate := time.Date(2023, time.July, 25, 11, 20, 0, 0, time.UTC)
				expectedResponse := &blnkgo.Transaction{
					ParentTransaction: blnkgo.ParentTransaction{
						Amount:        1000,
						Reference:     "ref-21",
						Precision:     100,
						Currency:      "USD",
						Source:        "@bank-account",
						Destination:   "@World",
						Status:        blnkgo.PryTransactionStatusApplied,
						Description:   "Test Transaction",
						EffectiveDate: &effectiveDate,
					},
					TransactionID: "tx-123",
					CreatedAt:     fixedTime,
				}

				m.On("NewRequest", "transactions/tx-123", http.MethodGet, nil).
					Return(&http.Request{}, nil)
				m.On("CallWithRetry", mock.Anything, mock.Anything).
					Return(&http.Response{StatusCode: http.StatusOK}, nil).
					Run(func(args mock.Arguments) {
						transaction := args.Get(1).(*blnkgo.Transaction)
						*transaction = *expectedResponse
					})
			},
		},
		{
			name:        "empty transaction ID",
			id:          "",
			expectError: true,
			errorMsg:    "transactionID is required",
			setupMocks:  func(m *MockClient) {},
		},
		{
			name:        "request creation failure",
			id:          "tx-123",
			expectError: true,
			errorMsg:    "failed to create request",
			setupMocks: func(m *MockClient) {
				m.On("NewRequest", "transactions/tx-123", http.MethodGet, nil).
					Return(nil, errors.New("failed to create request"))
			},
		},
		{
			name:        "server error",
			id:          "tx-123",
			expectError: true,
			errorMsg:    "server error",
			statusCode:  http.StatusInternalServerError,
			setupMocks: func(m *MockClient) {
				m.On("NewRequest", "transactions/tx-123", http.MethodGet, nil).
					Return(&http.Request{}, nil)
				m.On("CallWithRetry", mock.Anything, mock.Anything).
					Return(&http.Response{StatusCode: http.StatusInternalServerError},
						errors.New("server error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient, svc := setupTransactionService()
			tt.setupMocks(mockClient)

			transaction, resp, err := svc.Get(tt.id)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				if tt.id == "" {
					assert.Nil(t, resp)
					mockClient.AssertNotCalled(t, "NewRequest")
					mockClient.AssertNotCalled(t, "CallWithRetry")
				} else if tt.name == "request creation failure" {
					assert.Nil(t, transaction)
					assert.Nil(t, resp)
					mockClient.AssertNotCalled(t, "CallWithRetry")
				} else {
					assert.Equal(t, tt.statusCode, resp.StatusCode)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, transaction)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.statusCode, resp.StatusCode)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestTransactionService_Filter_Success(t *testing.T) {
	mockClient, svc := setupTransactionService()

	body := blnkgo.FilterParams{
		Filters: []blnkgo.Filter{
			{Field: "status", Operator: blnkgo.OpEqual, Value: "APPLIED"},
			{Field: "amount", Operator: blnkgo.OpGreaterThanOrEqual, Value: 10000},
		},
		SortBy:    "amount",
		SortOrder: "desc",
		Limit:     50,
	}

	fixedTime := time.Date(2024, time.January, 15, 10, 30, 0, 0, time.UTC)
	expectedResponse := &blnkgo.FilterResponse{
		Data: []blnkgo.Transaction{
			{
				ParentTransaction: blnkgo.ParentTransaction{
					Amount:   15000,
					Currency: "USD",
					Status:   blnkgo.PryTransactionStatusApplied,
				},
				TransactionID: "txn_abc123",
				CreatedAt:     fixedTime,
			},
		},
	}

	mockClient.On("NewRequest", "transactions/filter", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		result := args.Get(1).(*blnkgo.FilterResponse)
		*result = *expectedResponse
	})

	result, resp, err := svc.Filter(body)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_Filter_DateRange(t *testing.T) {
	mockClient, svc := setupTransactionService()

	body := blnkgo.FilterParams{
		Filters: []blnkgo.Filter{
			{Field: "created_at", Operator: blnkgo.OpGreaterThanOrEqual, Value: "2024-01-01"},
			{Field: "created_at", Operator: blnkgo.OpLessThan, Value: "2024-04-01"},
		},
	}

	fixedTime := time.Date(2024, time.February, 14, 9, 0, 0, 0, time.UTC)
	expectedResponse := &blnkgo.FilterResponse{
		Data: []blnkgo.Transaction{
			{
				ParentTransaction: blnkgo.ParentTransaction{
					Amount:   5000,
					Currency: "USD",
					Status:   blnkgo.PryTransactionStatusApplied,
				},
				TransactionID: "txn_q1_001",
				CreatedAt:     fixedTime,
			},
		},
	}

	mockClient.On("NewRequest", "transactions/filter", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		result := args.Get(1).(*blnkgo.FilterResponse)
		*result = *expectedResponse
	})

	result, resp, err := svc.Filter(body)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, resp)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_Filter_NewRequestError(t *testing.T) {
	mockClient, svc := setupTransactionService()

	body := blnkgo.FilterParams{
		Filters: []blnkgo.Filter{
			{Field: "status", Operator: blnkgo.OpEqual, Value: "APPLIED"},
		},
	}

	mockClient.On("NewRequest", "transactions/filter", http.MethodPost, body).Return(nil, fmt.Errorf("request error"))

	result, resp, err := svc.Filter(body)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Nil(t, resp)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_Filter_ServerError(t *testing.T) {
	mockClient, svc := setupTransactionService()

	body := blnkgo.FilterParams{
		Filters: []blnkgo.Filter{
			{Field: "status", Operator: blnkgo.OpEqual, Value: "APPLIED"},
		},
	}

	expectedResp := &http.Response{StatusCode: http.StatusInternalServerError}

	mockClient.On("NewRequest", "transactions/filter", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(expectedResp, fmt.Errorf("server error"))

	result, resp, err := svc.Filter(body)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestCreateTransactionWithInflightCommitDate(t *testing.T) {
	mockClient, svc := setupTransactionService()
	effectiveDate := time.Date(2023, time.September, 20, 14, 30, 0, 0, time.UTC)
	inflightCommitDate := time.Date(2023, time.September, 25, 10, 0, 0, 0, time.UTC)

	body := blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      1000,
			Reference:   "ref-inflight-commit",
			Precision:   100,
			Currency:    "USD",
			Source:      "@bank-account",
			Destination: "@World",
			MetaData: map[string]interface{}{
				"transaction_type": "deposit",
				"customer_name":    "Bob Smith",
				"customer_id":      "bob-1234",
			},
			Description:   "Inflight Commit Test",
			EffectiveDate: &effectiveDate,
		},
		Inflight:           true,
		InflightCommitDate: &inflightCommitDate,
	}
	fixedTime := time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC)

	mockClient.On("NewRequest", "transactions", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusCreated}, nil).Run(func(args mock.Arguments) {
		transaction := args.Get(1).(*blnkgo.Transaction)
		*transaction = blnkgo.Transaction{
			ParentTransaction: body.ParentTransaction,
			TransactionID:     "txn-inflight-commit-123",
			CreatedAt:         fixedTime,
		}
	})

	transaction, resp, err := svc.Create(body)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "txn-inflight-commit-123", transaction.TransactionID)
	assert.Equal(t, fixedTime, transaction.CreatedAt)

	mockClient.AssertExpectations(t)
}

func TestCreateTransactionWithInflightExpiryAndCommitDate(t *testing.T) {
	mockClient, svc := setupTransactionService()
	effectiveDate := time.Date(2023, time.September, 20, 14, 30, 0, 0, time.UTC)
	inflightExpiryDate := time.Date(2023, time.September, 30, 23, 59, 59, 0, time.UTC)
	inflightCommitDate := time.Date(2023, time.September, 25, 10, 0, 0, 0, time.UTC)

	body := blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      2500,
			Reference:   "ref-both-dates",
			Precision:   100,
			Currency:    "EUR",
			Source:      "@escrow-account",
			Destination: "@merchant",
			MetaData: map[string]interface{}{
				"transaction_type": "escrow_release",
				"order_id":         "order-9876",
			},
			Description:   "Escrow with both dates",
			EffectiveDate: &effectiveDate,
		},
		Inflight:           true,
		InflightExpiryDate: &inflightExpiryDate,
		InflightCommitDate: &inflightCommitDate,
	}
	fixedTime := time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC)

	mockClient.On("NewRequest", "transactions", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusCreated}, nil).Run(func(args mock.Arguments) {
		transaction := args.Get(1).(*blnkgo.Transaction)
		*transaction = blnkgo.Transaction{
			ParentTransaction: body.ParentTransaction,
			TransactionID:     "txn-both-dates-456",
			CreatedAt:         fixedTime,
		}
	})

	transaction, resp, err := svc.Create(body)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "txn-both-dates-456", transaction.TransactionID)
	assert.Equal(t, fixedTime, transaction.CreatedAt)

	mockClient.AssertExpectations(t)
}

func TestCreateTransactionInflightCommitDateWithoutInflight(t *testing.T) {
	mockClient, svc := setupTransactionService()
	inflightCommitDate := time.Date(2023, time.September, 25, 10, 0, 0, 0, time.UTC)

	body := blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      500,
			Reference:   "ref-commit-no-inflight",
			Precision:   100,
			Currency:    "USD",
			Source:      "@bank-account",
			Destination: "@World",
			Description: "Commit date without inflight flag",
		},
		Inflight:           false,
		InflightCommitDate: &inflightCommitDate,
	}
	fixedTime := time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC)

	mockClient.On("NewRequest", "transactions", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusCreated}, nil).Run(func(args mock.Arguments) {
		transaction := args.Get(1).(*blnkgo.Transaction)
		*transaction = blnkgo.Transaction{
			ParentTransaction: body.ParentTransaction,
			TransactionID:     "txn-no-inflight-789",
			CreatedAt:         fixedTime,
		}
	})

	transaction, resp, err := svc.Create(body)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "txn-no-inflight-789", transaction.TransactionID)

	mockClient.AssertExpectations(t)
}
