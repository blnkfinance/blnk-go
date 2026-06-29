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

func setupBalanceMonitorService() (*MockClient, *blnkgo.BalanceMonitorService) {
	mockClient := &MockClient{}
	svc := blnkgo.NewBalanceMonitorService(mockClient)
	return mockClient, svc
}

func TestBalanceMonitorService_Create_Success(t *testing.T) {
	mockClient, svc := setupBalanceMonitorService()

	data := blnkgo.MonitorData{
		Condition: blnkgo.MonitorCondition{
			Field:     "balance",
			Operator:  blnkgo.MonitorConditionOperators("greater_than"),
			Value:     1000,
			Precision: 2,
		},
		Description: "Monitor balance",
		CallBackURL: "http://callback.url",
		BalanceID:   "balance-123",
	}

	expectedResp := &blnkgo.MonitorDataResp{
		MonitorData: data,
		MonitorID:   "monitor-123",
		CreatedAt:   time.Now().Format(time.RFC3339),
	}

	mockClient.On("NewRequest", "balance-monitors", http.MethodPost, data).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusCreated}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.MonitorDataResp)
		*resp = *expectedResp
	})

	resp, httpResp, err := svc.Create(data)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusCreated, httpResp.StatusCode)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestBalanceMonitorService_Create_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupBalanceMonitorService()

	data := blnkgo.MonitorData{
		Condition: blnkgo.MonitorCondition{
			Field:     "balance",
			Operator:  blnkgo.MonitorConditionOperators("greater_than"),
			Value:     1000,
			Precision: 2,
		},
		Description: "Monitor balance",
		BalanceID:   "balance-123",
		CallBackURL: "http://callback.url",
	}

	mockClient.On("NewRequest", "balance-monitors", http.MethodPost, data).Return(nil, errors.New("failed to create request"))

	resp, httpResp, err := svc.Create(data)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, httpResp)
	assert.Contains(t, err.Error(), "failed to create request")
	mockClient.AssertExpectations(t)
}

func TestBalanceMonitorService_Create_ServerError(t *testing.T) {
	mockClient, svc := setupBalanceMonitorService()

	data := blnkgo.MonitorData{
		Condition: blnkgo.MonitorCondition{
			Field:     "balance",
			Operator:  blnkgo.MonitorConditionOperators("greater_than"),
			Value:     1000,
			Precision: 2,
		},
		Description: "Monitor balance",
		BalanceID:   "balance-123",
		CallBackURL: "http://callback.url",
	}

	mockClient.On("NewRequest", "balance-monitors", http.MethodPost, data).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusInternalServerError}, errors.New("server error"))

	resp, httpResp, err := svc.Create(data)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusInternalServerError, httpResp.StatusCode)
	assert.Contains(t, err.Error(), "server error")
	mockClient.AssertExpectations(t)
}
func TestBalanceMonitorService_Get_Success(t *testing.T) {
	mockClient, svc := setupBalanceMonitorService()

	monitorID := "monitor-123"
	expectedResp := &blnkgo.MonitorDataResp{
		MonitorID:   monitorID,
		CreatedAt:   time.Now().Format(time.RFC3339),
		MonitorData: blnkgo.MonitorData{BalanceID: "balance-123"},
	}

	mockClient.On("NewRequest", "balance-monitors/"+monitorID, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.MonitorDataResp)
		*resp = *expectedResp
	})

	resp, httpResp, err := svc.Get(monitorID)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestBalanceMonitorService_Get_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupBalanceMonitorService()

	monitorID := "monitor-123"

	mockClient.On("NewRequest", "balance-monitors/"+monitorID, http.MethodGet, nil).Return(nil, errors.New("failed to create request"))

	resp, httpResp, err := svc.Get(monitorID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, httpResp)
	assert.Contains(t, err.Error(), "failed to create request")
	mockClient.AssertExpectations(t)
}

func TestBalanceMonitorService_Get_ServerError(t *testing.T) {
	mockClient, svc := setupBalanceMonitorService()

	monitorID := "monitor-123"

	mockClient.On("NewRequest", "balance-monitors/"+monitorID, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusInternalServerError}, errors.New("server error"))

	resp, httpResp, err := svc.Get(monitorID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusInternalServerError, httpResp.StatusCode)
	assert.Contains(t, err.Error(), "server error")
	mockClient.AssertExpectations(t)
}
func TestBalanceMonitorService_List_Success(t *testing.T) {
	mockClient, svc := setupBalanceMonitorService()

	expectedResp := []blnkgo.MonitorDataResp{
		{
			MonitorID:   "monitor-123",
			CreatedAt:   time.Now().Format(time.RFC3339),
			MonitorData: blnkgo.MonitorData{BalanceID: "balance-123"},
		},
		{
			MonitorID:   "monitor-456",
			CreatedAt:   time.Now().Format(time.RFC3339),
			MonitorData: blnkgo.MonitorData{BalanceID: "balance-456"},
		},
	}

	mockClient.On("NewRequest", "balance-monitors", http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*[]blnkgo.MonitorDataResp)
		*resp = expectedResp
	})

	resp, httpResp, err := svc.List()

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestBalanceMonitorService_List_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupBalanceMonitorService()

	mockClient.On("NewRequest", "balance-monitors", http.MethodGet, nil).Return(nil, errors.New("failed to create request"))

	resp, httpResp, err := svc.List()

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, httpResp)
	assert.Contains(t, err.Error(), "failed to create request")
	mockClient.AssertExpectations(t)
}

func TestBalanceMonitorService_List_ServerError(t *testing.T) {
	mockClient, svc := setupBalanceMonitorService()

	mockClient.On("NewRequest", "balance-monitors", http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusInternalServerError}, errors.New("server error"))

	resp, httpResp, err := svc.List()

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusInternalServerError, httpResp.StatusCode)
	assert.Contains(t, err.Error(), "server error")
	mockClient.AssertExpectations(t)
}
func TestBalanceMonitorService_Update_Success(t *testing.T) {
	mockClient, svc := setupBalanceMonitorService()

	monitorID := "monitor-123"
	data := blnkgo.MonitorData{
		Condition: blnkgo.MonitorCondition{
			Field:     "balance",
			Operator:  blnkgo.MonitorConditionOperators("greater_than"),
			Value:     1000,
			Precision: 2,
		},
		Description: "Monitor balance",
		CallBackURL: "http://callback.url",
		BalanceID:   "balance-123",
	}

	expectedResp := &blnkgo.MonitorDataResp{
		MonitorData: data,
		MonitorID:   monitorID,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}

	mockClient.On("NewRequest", "balance-monitors/"+monitorID, http.MethodPut, data).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.MonitorDataResp)
		*resp = *expectedResp
	})

	resp, httpResp, err := svc.Update(monitorID, data)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestBalanceMonitorService_Update_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupBalanceMonitorService()

	monitorID := "monitor-123"
	data := blnkgo.MonitorData{
		Condition: blnkgo.MonitorCondition{
			Field:     "balance",
			Operator:  blnkgo.MonitorConditionOperators("greater_than"),
			Value:     1000,
			Precision: 2,
		},
		Description: "Monitor balance",
		CallBackURL: "http://callback.url",
		BalanceID:   "balance-123",
	}

	mockClient.On("NewRequest", "balance-monitors/"+monitorID, http.MethodPut, data).Return(nil, errors.New("failed to create request"))

	resp, httpResp, err := svc.Update(monitorID, data)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, httpResp)
	assert.Contains(t, err.Error(), "failed to create request")
	mockClient.AssertExpectations(t)
}

func TestBalanceMonitorService_Update_ServerError(t *testing.T) {
	mockClient, svc := setupBalanceMonitorService()

	monitorID := "monitor-123"
	data := blnkgo.MonitorData{
		Condition: blnkgo.MonitorCondition{
			Field:     "balance",
			Operator:  blnkgo.MonitorConditionOperators("greater_than"),
			Value:     1000,
			Precision: 2,
		},
		Description: "Monitor balance",
		CallBackURL: "http://callback.url",
		BalanceID:   "balance-123",
	}

	mockClient.On("NewRequest", "balance-monitors/"+monitorID, http.MethodPut, data).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusInternalServerError}, errors.New("server error"))

	resp, httpResp, err := svc.Update(monitorID, data)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusInternalServerError, httpResp.StatusCode)
	assert.Contains(t, err.Error(), "server error")
	mockClient.AssertExpectations(t)
}

func TestBalanceMonitorService_Delete_Success(t *testing.T) {
	mockClient, svc := setupBalanceMonitorService()

	monitorID := "mon_test_123"
	path := "balance-monitors/" + monitorID

	mockClient.On("NewRequest", path, http.MethodDelete, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.DeleteBalanceMonitorResponse)
		*resp = blnkgo.DeleteBalanceMonitorResponse{Message: "BalanceMonitor deleted successfully"}
	})

	deleted, httpResp, err := svc.Delete(monitorID)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, "BalanceMonitor deleted successfully", deleted.Message)
	mockClient.AssertExpectations(t)
}

func TestBalanceMonitorService_Delete_ValidationError(t *testing.T) {
	mockClient, svc := setupBalanceMonitorService()

	_, _, err := svc.Delete("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "monitor id is required")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestBalanceMonitorService_Delete_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupBalanceMonitorService()

	monitorID := "mon_test_123"
	path := "balance-monitors/" + monitorID

	mockClient.On("NewRequest", path, http.MethodDelete, nil).Return(nil, errors.New("failed to create request"))

	deleted, httpResp, err := svc.Delete(monitorID)

	assert.Error(t, err)
	assert.Nil(t, deleted)
	assert.Nil(t, httpResp)
	mockClient.AssertExpectations(t)
}
