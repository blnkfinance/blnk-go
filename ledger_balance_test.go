package blnkgo_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupLedgerBalanceService() (*MockClient, *blnkgo.LedgerBalanceService) {
	mockClient := &MockClient{}
	svc := blnkgo.NewLedgerBalanceService(mockClient)
	return mockClient, svc
}

func TestLedgerBalanceService_Create_Success(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()

	body := blnkgo.CreateLedgerBalanceRequest{
		LedgerID:   "ledger123",
		IdentityID: "identity123",
		Currency:   "USD",
	}

	fixedTime := time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC)
	expectedResponse := &blnkgo.LedgerBalance{
		BalanceID:  "balance123",
		LedgerID:   body.LedgerID,
		IdentityID: body.IdentityID,
		Currency:   body.Currency,
		CreatedAt:  fixedTime,
	}

	mockClient.On("NewRequest", "balances", http.MethodPost, body).Return(&http.Request{}, nil)

	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusCreated,
	}, nil).Run(func(args mock.Arguments) {
		ledgerBalance := args.Get(1).(*blnkgo.LedgerBalance)
		*ledgerBalance = *expectedResponse
	})

	ledgerBalance, resp, err := svc.Create(body)

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, ledgerBalance)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_Create_EmptyRequest(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	body := blnkgo.CreateLedgerBalanceRequest{}

	mockClient.On("NewRequest", "balances", http.MethodPost, body).Return(nil, fmt.Errorf("invalid request"))
	ledgerBalance, resp, err := svc.Create(body)

	assert.Error(t, err)
	assert.Nil(t, ledgerBalance)
	assert.Nil(t, resp)
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_Create_ServerError(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	body := blnkgo.CreateLedgerBalanceRequest{
		LedgerID:   "ledger123",
		IdentityID: "identity123",
		Currency:   "USD",
	}

	expectedResp := &http.Response{StatusCode: http.StatusInternalServerError}

	mockClient.On("NewRequest", "balances", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(expectedResp, fmt.Errorf("server error"))

	ledgerBalance, resp, err := svc.Create(body)

	assert.Error(t, err)
	assert.Nil(t, ledgerBalance)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_Get_Success(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	balanceID := "balance123"

	fixedTime := time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC)
	expectedResponse := &blnkgo.LedgerBalance{
		BalanceID:  balanceID,
		LedgerID:   "ledger123",
		IdentityID: "identity123",
		Currency:   "USD",
		CreatedAt:  fixedTime,
	}

	mockClient.On("NewRequest", fmt.Sprintf("balances/%s", balanceID), http.MethodGet, nil).Return(&http.Request{}, nil)

	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		ledgerBalance := args.Get(1).(*blnkgo.LedgerBalance)
		*ledgerBalance = *expectedResponse
	})
	ledgerBalance, resp, err := svc.Get(balanceID)

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, ledgerBalance)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_Get_CorrectEndpoint(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	balanceID := "balance123"
	expectedEndpoint := fmt.Sprintf("balances/%s", balanceID)
	mockClient.On("NewRequest", expectedEndpoint, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil)
	_, _, _ = svc.Get(balanceID)

	mockClient.AssertCalled(t, "NewRequest", expectedEndpoint, http.MethodGet, nil)
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_Get_EmptyID(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	balanceID := ""

	ledgerBalance, resp, err := svc.Get(balanceID)

	assert.Error(t, err)
	assert.Nil(t, ledgerBalance)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
	mockClient.AssertNotCalled(t, "CallWithRetry")
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_GetByIndicator(t *testing.T) {
	tests := []struct {
		name        string
		indicator   string
		currency    string
		expectError bool
		errorMsg    string
		statusCode  int
		setupMocks  func(*MockClient, string, string)
	}{
		{
			name:        "successful get by indicator",
			indicator:   "credit",
			currency:    "USD",
			expectError: false,
			statusCode:  http.StatusOK,
			setupMocks: func(m *MockClient, indicator, currency string) {
				fixedTime := time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC)
				expectedResponse := &blnkgo.LedgerBalance{
					BalanceID:  "balance123",
					LedgerID:   "ledger123",
					IdentityID: "identity123",
					Currency:   currency,
					Indicator:  indicator,
					CreatedAt:  fixedTime,
				}
				endpoint := fmt.Sprintf("balances/indicator/%s/currency/%s", indicator, currency)
				m.On("NewRequest", endpoint, http.MethodGet, nil).
					Return(&http.Request{}, nil)
				m.On("CallWithRetry", mock.Anything, mock.Anything).
					Return(&http.Response{StatusCode: http.StatusOK}, nil).
					Run(func(args mock.Arguments) {
						ledgerBalance := args.Get(1).(*blnkgo.LedgerBalance)
						*ledgerBalance = *expectedResponse
					})
			},
		},
		{
			name:        "empty indicator",
			indicator:   "",
			currency:    "USD",
			expectError: true,
			errorMsg:    "indicator is required",
			setupMocks:  func(m *MockClient, indicator, currency string) {},
		},
		{
			name:        "empty currency",
			indicator:   "debit",
			currency:    "",
			expectError: true,
			errorMsg:    "currency is required",
			setupMocks:  func(m *MockClient, indicator, currency string) {},
		},
		{
			name:        "request creation failure",
			indicator:   "credit",
			currency:    "EUR",
			expectError: true,
			errorMsg:    "failed to create request",
			setupMocks: func(m *MockClient, indicator, currency string) {
				endpoint := fmt.Sprintf("balances/indicator/%s/currency/%s", indicator, currency)
				m.On("NewRequest", endpoint, http.MethodGet, nil).
					Return(nil, fmt.Errorf("failed to create request"))
			},
		},
		{
			name:        "server error",
			indicator:   "credit",
			currency:    "GBP",
			expectError: true,
			errorMsg:    "server error",
			statusCode:  http.StatusInternalServerError,
			setupMocks: func(m *MockClient, indicator, currency string) {
				endpoint := fmt.Sprintf("balances/indicator/%s/currency/%s", indicator, currency)
				m.On("NewRequest", endpoint, http.MethodGet, nil).
					Return(&http.Request{}, nil)
				m.On("CallWithRetry", mock.Anything, mock.Anything).
					Return(&http.Response{StatusCode: http.StatusInternalServerError},
						fmt.Errorf("server error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient, svc := setupLedgerBalanceService()
			tt.setupMocks(mockClient, tt.indicator, tt.currency)

			ledgerBalance, resp, err := svc.GetByIndicator(tt.indicator, tt.currency)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				if tt.indicator == "" || tt.currency == "" {
					assert.Nil(t, resp)
					mockClient.AssertNotCalled(t, "NewRequest")
					mockClient.AssertNotCalled(t, "CallWithRetry")
				} else if tt.name == "request creation failure" {
					assert.Nil(t, ledgerBalance)
					assert.Nil(t, resp)
					mockClient.AssertNotCalled(t, "CallWithRetry")
				} else {
					assert.Equal(t, tt.statusCode, resp.StatusCode)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ledgerBalance)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.statusCode, resp.StatusCode)
				assert.Equal(t, tt.indicator, ledgerBalance.Indicator)
				assert.Equal(t, tt.currency, ledgerBalance.Currency)
			}
			mockClient.AssertExpectations(t)
		})
	}
}
