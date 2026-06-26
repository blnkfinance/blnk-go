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

func setupReconciliationService() (*MockClient, *blnkgo.ReconciliationService) {
	mockClient := &MockClient{}
	svc := blnkgo.NewReconciliationService(mockClient)
	return mockClient, svc
}

func TestReconciliationService_CreateMatchingRule_Success(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	matcher := blnkgo.Matcher{
		Name:        "Test Matcher",
		Description: "Test Description",
		Criteria: []blnkgo.Criteria{
			{
				Field:    "amount",
				Operator: "equals",
			},
		},
	}

	expectedResp := &blnkgo.RunReconResp{
		Matcher:   matcher,
		RuleID:    "rule-123",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	mockClient.On("NewRequest", "reconciliation/matching-rules", http.MethodPost, matcher).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusCreated}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.RunReconResp)
		*resp = *expectedResp
	})

	resp, httpResp, err := svc.CreateMatchingRule(matcher)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusCreated, httpResp.StatusCode)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_CreateMatchingRule_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	matcher := blnkgo.Matcher{
		Name:        "Test Matcher",
		Description: "Test Description",
		Criteria: []blnkgo.Criteria{
			{
				Field:    "amount",
				Operator: "equals",
			},
		},
	}

	mockClient.On("NewRequest", "reconciliation/matching-rules", http.MethodPost, matcher).Return(nil, errors.New("failed to create request"))

	resp, httpResp, err := svc.CreateMatchingRule(matcher)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, httpResp)
	assert.Contains(t, err.Error(), "failed to create request")
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_CreateMatchingRule_ServerError(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	matcher := blnkgo.Matcher{
		Name:        "Test Matcher",
		Description: "Test Description",
		Criteria: []blnkgo.Criteria{
			{
				Field:    "amount",
				Operator: "equals",
			},
		},
	}

	mockClient.On("NewRequest", "reconciliation/matching-rules", http.MethodPost, matcher).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusInternalServerError}, errors.New("server error"))

	resp, httpResp, err := svc.CreateMatchingRule(matcher)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusInternalServerError, httpResp.StatusCode)
	assert.Contains(t, err.Error(), "server error")
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_Run_Success(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	data := blnkgo.RunReconData{
		UploadID:         "upload-123",
		Strategy:         "default",
		DryRun:           true,
		GroupingCriteria: "amount",
		MatchingRuleIDs:  []string{"rule-123"},
	}

	expectedResp := &blnkgo.RunReconResp{
		RuleID:    "rule-123",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	mockClient.On("NewRequest", "reconciliation/start", http.MethodPost, data).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.RunReconResp)
		*resp = *expectedResp
	})

	resp, httpResp, err := svc.Run(data)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_Run_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	data := blnkgo.RunReconData{
		UploadID:         "upload-123",
		Strategy:         "default",
		DryRun:           true,
		GroupingCriteria: "amount",
		MatchingRuleIDs:  []string{"rule-123"},
	}

	mockClient.On("NewRequest", "reconciliation/start", http.MethodPost, data).Return(nil, errors.New("failed to create request"))

	resp, httpResp, err := svc.Run(data)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, httpResp)
	assert.Contains(t, err.Error(), "failed to create request")
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_Run_ServerError(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	data := blnkgo.RunReconData{
		UploadID:         "upload-123",
		Strategy:         "default",
		DryRun:           true,
		GroupingCriteria: "amount",
		MatchingRuleIDs:  []string{"rule-123"},
	}

	mockClient.On("NewRequest", "reconciliation/start", http.MethodPost, data).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusInternalServerError}, errors.New("server error"))

	resp, httpResp, err := svc.Run(data)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusInternalServerError, httpResp.StatusCode)
	assert.Contains(t, err.Error(), "server error")
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_Upload_Success(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	source := "test-source"
	file := []byte("test file content")
	fileName := "testfile.txt"

	expectedResp := &blnkgo.ReconciliationUploadResp{
		UploadID:    "upload-123",
		RecordCount: 100,
		Source:      source,
	}

	mockClient.On("NewFileUploadRequest", "reconciliation/upload", "file", file, fileName, map[string]string{"source": source}).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusCreated}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.ReconciliationUploadResp)
		*resp = *expectedResp
	})

	resp, httpResp, err := svc.Upload(source, file, fileName)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusCreated, httpResp.StatusCode)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_Upload_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	source := "test-source"
	file := []byte("test file content")
	fileName := "testfile.txt"

	mockClient.On("NewFileUploadRequest", "reconciliation/upload", "file", file, fileName, map[string]string{"source": source}).Return(nil, errors.New("failed to create request"))

	resp, httpResp, err := svc.Upload(source, file, fileName)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, httpResp)
	assert.Contains(t, err.Error(), "failed to create request")
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_Upload_ServerError(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	source := "test-source"
	file := []byte("test file content")
	fileName := "testfile.txt"

	mockClient.On("NewFileUploadRequest", "reconciliation/upload", "file", file, fileName, map[string]string{"source": source}).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusInternalServerError}, errors.New("server error"))

	resp, httpResp, err := svc.Upload(source, file, fileName)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusInternalServerError, httpResp.StatusCode)
	assert.Contains(t, err.Error(), "server error")
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_RunInstant_Success(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	data := blnkgo.RunInstantReconData{
		ExternalTransactions: []blnkgo.ExternalTransaction{
			{
				ID:          "ext-1",
				Amount:      10.5,
				Reference:   "REF-1",
				Currency:    "USD",
				Description: "Test payment",
				Date:        time.Date(2024, 11, 15, 14, 25, 30, 0, time.UTC),
				Source:      "bank-api",
			},
		},
		Strategy:        blnkgo.ReconciliationStrategyOneToOne,
		DryRun:          true,
		MatchingRuleIDs: []string{"rule-123"},
	}

	expectedResp := &blnkgo.RunInstantReconResp{
		ReconciliationID: "recon_abc123",
	}

	mockClient.On("NewRequest", "reconciliation/start-instant", http.MethodPost, data).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.RunInstantReconResp)
		*resp = *expectedResp
	})

	resp, httpResp, err := svc.RunInstant(data)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_RunInstant_ValidationError(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	data := blnkgo.RunInstantReconData{
		ExternalTransactions: nil,
		Strategy:             blnkgo.ReconciliationStrategyOneToOne,
		MatchingRuleIDs:      []string{"rule-123"},
	}

	resp, httpResp, err := svc.RunInstant(data)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, httpResp)
	assert.Contains(t, err.Error(), "external_transactions must be a non-empty array")
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_RunInstant_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	data := blnkgo.RunInstantReconData{
		ExternalTransactions: []blnkgo.ExternalTransaction{
			{
				ID:          "ext-1",
				Amount:      10.5,
				Reference:   "REF-1",
				Currency:    "USD",
				Description: "Test payment",
				Date:        time.Date(2024, 11, 15, 14, 25, 30, 0, time.UTC),
				Source:      "bank-api",
			},
		},
		Strategy:        blnkgo.ReconciliationStrategyOneToOne,
		MatchingRuleIDs: []string{"rule-123"},
	}

	mockClient.On("NewRequest", "reconciliation/start-instant", http.MethodPost, data).Return(nil, errors.New("failed to create request"))

	resp, httpResp, err := svc.RunInstant(data)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, httpResp)
	assert.Contains(t, err.Error(), "failed to create request")
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_RunInstant_ServerError(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	data := blnkgo.RunInstantReconData{
		ExternalTransactions: []blnkgo.ExternalTransaction{
			{
				ID:          "ext-1",
				Amount:      10.5,
				Reference:   "REF-1",
				Currency:    "USD",
				Description: "Test payment",
				Date:        time.Date(2024, 11, 15, 14, 25, 30, 0, time.UTC),
				Source:      "bank-api",
			},
		},
		Strategy:        blnkgo.ReconciliationStrategyOneToOne,
		MatchingRuleIDs: []string{"rule-123"},
	}

	mockClient.On("NewRequest", "reconciliation/start-instant", http.MethodPost, data).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusInternalServerError}, errors.New("server error"))

	resp, httpResp, err := svc.RunInstant(data)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusInternalServerError, httpResp.StatusCode)
	assert.Contains(t, err.Error(), "server error")
	mockClient.AssertExpectations(t)
}
