package blnkgo_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func timePtr(t time.Time) *time.Time {
	return &t
}

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

	expectedResp := &blnkgo.StartReconciliationResponse{
		ReconciliationID: "recon_test_123",
	}

	mockClient.On("NewRequest", "reconciliation/start", http.MethodPost, data).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.StartReconciliationResponse)
		*resp = *expectedResp
	})

	resp, httpResp, err := svc.Run(data)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_Run_ResponseUnmarshalJSON(t *testing.T) {
	var resp blnkgo.StartReconciliationResponse
	err := json.Unmarshal([]byte(`{"reconciliation_id":"recon_abc123"}`), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "recon_abc123", resp.ReconciliationID)
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
				Date:        timePtr(time.Date(2024, 11, 15, 14, 25, 30, 0, time.UTC)),
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
				Date:        timePtr(time.Date(2024, 11, 15, 14, 25, 30, 0, time.UTC)),
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
				Date:        timePtr(time.Date(2024, 11, 15, 14, 25, 30, 0, time.UTC)),
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

func TestRunInstantReconData_MinimalJSONOmitsOptionalFields(t *testing.T) {
	data := blnkgo.RunInstantReconData{
		ExternalTransactions: []blnkgo.ExternalTransaction{
			{ID: "ext-1", Amount: 1, Reference: "r1", Currency: "USD"},
		},
		Strategy:        blnkgo.ReconciliationStrategyOneToOne,
		MatchingRuleIDs: []string{"rule_1"},
	}

	body, err := json.Marshal(data)
	assert.NoError(t, err)

	encoded := string(body)
	assert.NotContains(t, encoded, `"date"`)
	assert.NotContains(t, encoded, `"description"`)
	assert.NotContains(t, encoded, `"source"`)
	assert.Contains(t, encoded, `"id":"ext-1"`)
	assert.Contains(t, encoded, `"amount":1`)
	assert.Contains(t, encoded, `"reference":"r1"`)
	assert.Contains(t, encoded, `"currency":"USD"`)
}

func TestReconciliationService_Get_Success(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	reconID := "recon_abc123"
	startedAt := time.Date(2025, 3, 15, 17, 43, 26, 0, time.UTC)
	expected := &blnkgo.Reconciliation{
		ReconciliationID:      reconID,
		UploadID:              "instant_xyz",
		Status:                "completed",
		MatchedTransactions:   3,
		UnmatchedTransactions: 0,
		IsDryRun:              true,
		StartedAt:             startedAt,
	}

	mockClient.On("NewRequest", fmt.Sprintf("reconciliation/%s", reconID), http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		recon := args.Get(1).(*blnkgo.Reconciliation)
		*recon = *expected
	})

	recon, httpResp, err := svc.Get(reconID)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, expected, recon)
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_Get_EmptyID(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	recon, httpResp, err := svc.Get("")

	assert.Error(t, err)
	assert.Nil(t, recon)
	assert.Nil(t, httpResp)
	assert.Contains(t, err.Error(), "reconciliation id is required")
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_Get_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	reconID := "recon_abc123"
	mockClient.On("NewRequest", fmt.Sprintf("reconciliation/%s", reconID), http.MethodGet, nil).Return(nil, errors.New("failed to create request"))

	recon, httpResp, err := svc.Get(reconID)

	assert.Error(t, err)
	assert.Nil(t, recon)
	assert.Nil(t, httpResp)
	assert.Contains(t, err.Error(), "failed to create request")
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_Get_NotFound(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	reconID := "recon_missing"
	mockClient.On("NewRequest", fmt.Sprintf("reconciliation/%s", reconID), http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusNotFound}, errors.New("not found"))

	recon, httpResp, err := svc.Get(reconID)

	assert.Error(t, err)
	assert.Nil(t, recon)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusNotFound, httpResp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_UpdateMatchingRule_Success(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	ruleID := "rule_abc123"
	matcher := blnkgo.Matcher{
		Name:        "Updated Rule",
		Description: "Updated description",
		Criteria: []blnkgo.Criteria{
			{
				Field:    blnkgo.CriteriaFieldAmount,
				Operator: blnkgo.ReconciliationOperatorEquals,
			},
		},
	}
	expected := &blnkgo.RunReconResp{
		Matcher:   matcher,
		RuleID:    ruleID,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	mockClient.On("NewRequest", fmt.Sprintf("reconciliation/matching-rules/%s", ruleID), http.MethodPut, matcher).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.RunReconResp)
		*resp = *expected
	})

	result, httpResp, err := svc.UpdateMatchingRule(ruleID, matcher)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, expected, result)
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_UpdateMatchingRule_EmptyID(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	result, httpResp, err := svc.UpdateMatchingRule("", blnkgo.Matcher{Name: "x"})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Nil(t, httpResp)
	assert.Contains(t, err.Error(), "matching rule id is required")
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_UpdateMatchingRule_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	ruleID := "rule_abc123"
	matcher := blnkgo.Matcher{Name: "Updated Rule"}

	mockClient.On("NewRequest", fmt.Sprintf("reconciliation/matching-rules/%s", ruleID), http.MethodPut, matcher).Return(nil, errors.New("failed to create request"))

	result, httpResp, err := svc.UpdateMatchingRule(ruleID, matcher)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Nil(t, httpResp)
	assert.Contains(t, err.Error(), "failed to create request")
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_UpdateMatchingRule_NotFound(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	ruleID := "rule_missing"
	matcher := blnkgo.Matcher{Name: "Updated Rule"}

	mockClient.On("NewRequest", fmt.Sprintf("reconciliation/matching-rules/%s", ruleID), http.MethodPut, matcher).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusNotFound}, errors.New("not found"))

	result, httpResp, err := svc.UpdateMatchingRule(ruleID, matcher)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusNotFound, httpResp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_DeleteMatchingRule_Success(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	ruleID := "rule_abc123"
	expected := &blnkgo.DeleteMatchingRuleResp{
		Message: "Matching rule deleted successfully",
	}

	mockClient.On("NewRequest", fmt.Sprintf("reconciliation/matching-rules/%s", ruleID), http.MethodDelete, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.DeleteMatchingRuleResp)
		*resp = *expected
	})

	result, httpResp, err := svc.DeleteMatchingRule(ruleID)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, expected, result)
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_DeleteMatchingRule_EmptyID(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	result, httpResp, err := svc.DeleteMatchingRule("")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Nil(t, httpResp)
	assert.Contains(t, err.Error(), "matching rule id is required")
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_DeleteMatchingRule_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	ruleID := "rule_abc123"
	mockClient.On("NewRequest", fmt.Sprintf("reconciliation/matching-rules/%s", ruleID), http.MethodDelete, nil).Return(nil, errors.New("failed to create request"))

	result, httpResp, err := svc.DeleteMatchingRule(ruleID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Nil(t, httpResp)
	assert.Contains(t, err.Error(), "failed to create request")
	mockClient.AssertExpectations(t)
}

func TestReconciliationService_DeleteMatchingRule_NotFound(t *testing.T) {
	mockClient, svc := setupReconciliationService()

	ruleID := "rule_missing"
	mockClient.On("NewRequest", fmt.Sprintf("reconciliation/matching-rules/%s", ruleID), http.MethodDelete, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusNotFound}, errors.New("not found"))

	result, httpResp, err := svc.DeleteMatchingRule(ruleID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusNotFound, httpResp.StatusCode)
	mockClient.AssertExpectations(t)
}
