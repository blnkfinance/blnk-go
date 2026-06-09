package blnkgo_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransactionService_GetByReference_Success(t *testing.T) {
	mockClient, svc := setupTransactionService()
	reference := "ref_8d2ce2f0-0d75-4a91-9d43-2ad2c2e6b9ad"
	fixedTime := time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC)

	expectedResponse := &blnkgo.Transaction{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      1000,
			Reference:   reference,
			Precision:   100,
			Currency:    "USD",
			Source:      "@FundingPool",
			Destination: "@Recipient",
			Status:      blnkgo.PryTransactionStatusApplied,
			Description: "Test Transaction",
		},
		TransactionID: "txn_8d2ce2f0-0d75-4a91-9d43-2ad2c2e6b9ad",
		CreatedAt:     fixedTime,
	}

	endpoint := fmt.Sprintf("transactions/reference/%s", reference)
	mockClient.On("NewRequest", endpoint, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		transaction := args.Get(1).(*blnkgo.Transaction)
		*transaction = *expectedResponse
	})

	transaction, resp, err := svc.GetByReference(reference)

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, transaction)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_GetByReference_PathEscapesReference(t *testing.T) {
	mockClient, svc := setupTransactionService()
	reference := "ref/with space?query#hash%25"
	endpoint := fmt.Sprintf("transactions/reference/%s", url.PathEscape(reference))

	mockClient.On("NewRequest", endpoint, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil)

	_, _, _ = svc.GetByReference(reference)

	mockClient.AssertCalled(t, "NewRequest", endpoint, http.MethodGet, nil)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_GetByReference_CorrectEndpoint(t *testing.T) {
	mockClient, svc := setupTransactionService()
	reference := "ref_8d2ce2f0-0d75-4a91-9d43-2ad2c2e6b9ad"
	endpoint := fmt.Sprintf("transactions/reference/%s", reference)

	mockClient.On("NewRequest", endpoint, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil)

	_, _, _ = svc.GetByReference(reference)

	mockClient.AssertCalled(t, "NewRequest", endpoint, http.MethodGet, nil)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_GetByReference_EmptyReference(t *testing.T) {
	mockClient, svc := setupTransactionService()

	transaction, resp, err := svc.GetByReference("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reference is required")
	assert.Nil(t, transaction)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
	mockClient.AssertNotCalled(t, "CallWithRetry")
	mockClient.AssertExpectations(t)
}

func TestTransactionService_GetByReference_NewRequestError(t *testing.T) {
	mockClient, svc := setupTransactionService()
	reference := "ref_8d2ce2f0-0d75-4a91-9d43-2ad2c2e6b9ad"
	endpoint := fmt.Sprintf("transactions/reference/%s", reference)

	mockClient.On("NewRequest", endpoint, http.MethodGet, nil).Return(nil, fmt.Errorf("request error"))

	transaction, resp, err := svc.GetByReference(reference)

	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "CallWithRetry")
	mockClient.AssertExpectations(t)
}

func TestTransactionService_GetByReference_ServerError(t *testing.T) {
	mockClient, svc := setupTransactionService()
	reference := "ref_8d2ce2f0-0d75-4a91-9d43-2ad2c2e6b9ad"
	endpoint := fmt.Sprintf("transactions/reference/%s", reference)
	expectedResp := &http.Response{StatusCode: http.StatusInternalServerError}

	mockClient.On("NewRequest", endpoint, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(expectedResp, fmt.Errorf("server error"))

	transaction, resp, err := svc.GetByReference(reference)

	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}
