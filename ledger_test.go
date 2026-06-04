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

func setupLedgerService() (*MockClient, *blnkgo.LedgerService) {
	mockClient := &MockClient{}
	svc := blnkgo.NewLedgerService(mockClient)
	return mockClient, svc
}
func TestLedgerService_Create_Success(t *testing.T) {
	mockClient, svc := setupLedgerService()

	body := blnkgo.CreateLedgerRequest{
		Name: "Test Ledger",
	}

	fixedTime := time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC)
	expectedResponse := &blnkgo.Ledger{
		Name:      body.Name,
		LedgerID:  "123",
		CreatedAt: fixedTime,
	}

	mockClient.On("NewRequest", "ledgers", http.MethodPost, body).Return(&http.Request{}, nil)

	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusCreated,
	}, nil).Run(func(args mock.Arguments) {
		ledger := args.Get(1).(*blnkgo.Ledger)
		*ledger = *expectedResponse
	})

	ledger, resp, err := svc.Create(body)

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, ledger)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestLedgerService_Create_EmptyRequest(t *testing.T) {
	mockClient, svc := setupLedgerService()
	body := blnkgo.CreateLedgerRequest{}

	mockClient.On("NewRequest", "ledgers", http.MethodPost, body).Return(nil, fmt.Errorf("invalid request"))
	ledger, resp, err := svc.Create(body)

	assert.Error(t, err)
	assert.Nil(t, ledger)
	assert.Nil(t, resp)
	mockClient.AssertExpectations(t)
}

func TestLedgerService_Create_ServerError(t *testing.T) {
	mockClient, svc := setupLedgerService()
	body := blnkgo.CreateLedgerRequest{
		Name: "Test Ledger",
	}

	expectedResp := &http.Response{StatusCode: http.StatusInternalServerError}

	mockClient.On("NewRequest", "ledgers", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(expectedResp, fmt.Errorf("server error"))

	ledger, resp, err := svc.Create(body)

	assert.Error(t, err)
	assert.Nil(t, ledger)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestLedgerService_List_Success(t *testing.T) {
	mockClient, svc := setupLedgerService()

	ledgerID := "123"
	fixedTime := time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC)
	expectedResponse := []blnkgo.Ledger{
		{
			Name:      "Test Ledger",
			LedgerID:  ledgerID,
			CreatedAt: fixedTime,
		},
	}

	mockClient.On("NewRequest", "ledgers", http.MethodGet, nil).Return(&http.Request{}, nil)

	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		ledgers := args.Get(1).(*[]blnkgo.Ledger)
		*ledgers = expectedResponse
	})
	ledger, resp, err := svc.List()

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, ledger)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestLedgerService_Get_Success(t *testing.T) {
	mockClient, svc := setupLedgerService()
	ledgerID := "123"

	fixedTime := time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC)
	expectedResponse := &blnkgo.Ledger{
		Name:      "Test Ledger",
		LedgerID:  ledgerID,
		CreatedAt: fixedTime,
	}

	mockClient.On("NewRequest", fmt.Sprintf("ledgers/%s", ledgerID), http.MethodGet, nil).Return(&http.Request{}, nil)

	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		ledger := args.Get(1).(*blnkgo.Ledger)
		*ledger = *expectedResponse
	})
	ledger, resp, err := svc.Get(ledgerID)

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, ledger)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

// http request is made witht the correct url
func TestLedgerService_Get_CorrectEndpoint(t *testing.T) {
	mockClient, svc := setupLedgerService()
	ledgerID := "123"
	expectedEndpoint := fmt.Sprintf("ledgers/%s", ledgerID)
	mockClient.On("NewRequest", expectedEndpoint, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil)
	_, _, _ = svc.Get(ledgerID)

	mockClient.AssertCalled(t, "NewRequest", expectedEndpoint, http.MethodGet, nil)
	mockClient.AssertExpectations(t)

}

func TestLedgerSerice_Get_EmptyID(t *testing.T) {
	mockClient, svc := setupLedgerService()
	ledgerID := ""

	ledger, resp, err := svc.Get(ledgerID)

	assert.Error(t, err)
	assert.Nil(t, ledger)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
	mockClient.AssertNotCalled(t, "CallWithRetry")
	mockClient.AssertExpectations(t)
}

func TestLedgerService_Update_Success(t *testing.T) {
	mockClient, svc := setupLedgerService()
	ledgerID := "ldg_123"
	body := blnkgo.UpdateLedgerRequest{
		Name: "Updated Customer Savings Account",
	}

	fixedTime := time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC)
	expectedResponse := &blnkgo.Ledger{
		LedgerID:  ledgerID,
		Name:      body.Name,
		CreatedAt: fixedTime,
	}

	mockClient.On("NewRequest", fmt.Sprintf("ledgers/%s", ledgerID), http.MethodPut, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		ledger := args.Get(1).(*blnkgo.Ledger)
		*ledger = *expectedResponse
	})

	ledger, resp, err := svc.Update(ledgerID, body)

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, ledger)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestLedgerService_Update_CorrectEndpoint(t *testing.T) {
	mockClient, svc := setupLedgerService()
	ledgerID := "ldg_123"
	body := blnkgo.UpdateLedgerRequest{Name: "Updated Name"}
	expectedEndpoint := fmt.Sprintf("ledgers/%s", ledgerID)

	mockClient.On("NewRequest", expectedEndpoint, http.MethodPut, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil)

	_, _, _ = svc.Update(ledgerID, body)

	mockClient.AssertCalled(t, "NewRequest", expectedEndpoint, http.MethodPut, body)
	mockClient.AssertExpectations(t)
}

func TestLedgerService_Update_EmptyID(t *testing.T) {
	mockClient, svc := setupLedgerService()
	body := blnkgo.UpdateLedgerRequest{Name: "Updated Name"}

	ledger, resp, err := svc.Update("", body)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
	assert.Nil(t, ledger)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
	mockClient.AssertNotCalled(t, "CallWithRetry")
	mockClient.AssertExpectations(t)
}

func TestLedgerService_Update_EmptyName(t *testing.T) {
	mockClient, svc := setupLedgerService()
	body := blnkgo.UpdateLedgerRequest{}

	ledger, resp, err := svc.Update("ldg_123", body)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
	assert.Nil(t, ledger)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
	mockClient.AssertNotCalled(t, "CallWithRetry")
	mockClient.AssertExpectations(t)
}

func TestLedgerService_Update_NewRequestError(t *testing.T) {
	mockClient, svc := setupLedgerService()
	ledgerID := "ldg_123"
	body := blnkgo.UpdateLedgerRequest{Name: "Updated Name"}

	mockClient.On("NewRequest", fmt.Sprintf("ledgers/%s", ledgerID), http.MethodPut, body).
		Return(nil, fmt.Errorf("request error"))

	ledger, resp, err := svc.Update(ledgerID, body)

	assert.Error(t, err)
	assert.Nil(t, ledger)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "CallWithRetry")
	mockClient.AssertExpectations(t)
}

func TestLedgerService_Update_ServerError(t *testing.T) {
	mockClient, svc := setupLedgerService()
	ledgerID := "ldg_123"
	body := blnkgo.UpdateLedgerRequest{Name: "Updated Name"}
	expectedResp := &http.Response{StatusCode: http.StatusInternalServerError}

	mockClient.On("NewRequest", fmt.Sprintf("ledgers/%s", ledgerID), http.MethodPut, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(expectedResp, fmt.Errorf("server error"))

	ledger, resp, err := svc.Update(ledgerID, body)

	assert.Error(t, err)
	assert.Nil(t, ledger)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestLedgerService_Filter_Success(t *testing.T) {
	mockClient, svc := setupLedgerService()

	body := blnkgo.FilterParams{
		Filters: []blnkgo.Filter{
			{Field: "name", Operator: blnkgo.OpILike, Value: "%savings%"},
		},
		SortBy:    "name",
		SortOrder: "asc",
	}

	fixedTime := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	expectedResponse := &blnkgo.FilterResponse{
		Data: []blnkgo.Ledger{
			{
				LedgerID:  "ldg_savings_001",
				Name:      "Customer Savings Accounts",
				CreatedAt: fixedTime,
			},
			{
				LedgerID:  "ldg_savings_002",
				Name:      "High-Yield Savings",
				CreatedAt: fixedTime,
			},
		},
	}

	mockClient.On("NewRequest", "ledgers/filter", http.MethodPost, body).Return(&http.Request{}, nil)
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

func TestLedgerService_Filter_NewRequestError(t *testing.T) {
	mockClient, svc := setupLedgerService()

	body := blnkgo.FilterParams{
		Filters: []blnkgo.Filter{
			{Field: "name", Operator: blnkgo.OpEqual, Value: "test"},
		},
	}

	mockClient.On("NewRequest", "ledgers/filter", http.MethodPost, body).Return(nil, fmt.Errorf("request error"))

	result, resp, err := svc.Filter(body)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Nil(t, resp)
	mockClient.AssertExpectations(t)
}

func TestLedgerService_Filter_ServerError(t *testing.T) {
	mockClient, svc := setupLedgerService()

	body := blnkgo.FilterParams{
		Filters: []blnkgo.Filter{
			{Field: "name", Operator: blnkgo.OpEqual, Value: "test"},
		},
	}

	expectedResp := &http.Response{StatusCode: http.StatusInternalServerError}

	mockClient.On("NewRequest", "ledgers/filter", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(expectedResp, fmt.Errorf("server error"))

	result, resp, err := svc.Filter(body)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}
