package blnkgo_test

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestTransactionLineage_UnmarshalJSON verifies TransactionLineage decodes the API
// response contract from docs.blnkfinance.com/reference/get-transaction-lineage.
func TestTransactionLineage_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		payload string
	}{
		{
			name: "docs example with string amounts",
			payload: `{
				"transaction_id": "txn_8d2ce2f0-0d75-4a91-9d43-2ad2c2e6b9ad",
				"fund_allocation": [
					{ "provider": "stripe", "amount": "2500" }
				],
				"shadow_transactions": [
					{
						"transaction_id": "txn_shadow_123",
						"reference": "ref_002_release_stripe_0",
						"amount": 25,
						"precision": 100,
						"currency": "USD",
						"status": "APPLIED"
					}
				]
			}`,
		},
		{
			name: "core serialization with numeric amounts",
			payload: `{
				"transaction_id": "txn_8d2ce2f0-0d75-4a91-9d43-2ad2c2e6b9ad",
				"fund_allocation": [
					{ "provider": "stripe", "amount": 2500 },
					{ "provider": "paypal", "amount": 1000 }
				],
				"shadow_transactions": []
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lineage blnkgo.TransactionLineage
			err := json.Unmarshal([]byte(tt.payload), &lineage)
			assert.NoError(t, err)

			assert.Equal(t, "txn_8d2ce2f0-0d75-4a91-9d43-2ad2c2e6b9ad", lineage.TransactionID)
			assert.NotEmpty(t, lineage.FundAllocation)
			assert.Equal(t, "stripe", lineage.FundAllocation[0].Provider)
			assert.Equal(t, 0, big.NewInt(2500).Cmp(lineage.FundAllocation[0].Amount))

			if tt.name == "docs example with string amounts" {
				assert.Len(t, lineage.ShadowTransactions, 1)
				shadow := lineage.ShadowTransactions[0]
				assert.Equal(t, "txn_shadow_123", shadow.TransactionID)
				assert.Equal(t, "ref_002_release_stripe_0", shadow.Reference)
				assert.Equal(t, "USD", shadow.Currency)
				assert.Equal(t, blnkgo.PryTransactionStatusApplied, shadow.Status)
				assert.Equal(t, float64(25), shadow.Amount)
			} else {
				assert.Empty(t, lineage.ShadowTransactions)
				assert.Len(t, lineage.FundAllocation, 2)
				assert.Equal(t, "paypal", lineage.FundAllocation[1].Provider)
				assert.Equal(t, 0, big.NewInt(1000).Cmp(lineage.FundAllocation[1].Amount))
			}
		})
	}
}

func TestTransactionService_GetLineage_Success(t *testing.T) {
	mockClient, svc := setupTransactionService()
	transactionID := "txn_8d2ce2f0-0d75-4a91-9d43-2ad2c2e6b9ad"

	expectedResponse := &blnkgo.TransactionLineage{
		TransactionID: transactionID,
		FundAllocation: []blnkgo.LineageFundAllocation{
			{
				Provider: "stripe",
				Amount:   big.NewInt(2500),
			},
		},
		ShadowTransactions: []blnkgo.Transaction{
			{
				TransactionID: "txn_shadow_123",
				ParentTransaction: blnkgo.ParentTransaction{
					Reference: "ref_002_release_stripe_0",
					Currency:  "USD",
					Status:    blnkgo.PryTransactionStatusApplied,
				},
			},
		},
	}

	endpoint := fmt.Sprintf("transactions/%s/lineage", transactionID)
	mockClient.On("NewRequest", endpoint, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		lineage := args.Get(1).(*blnkgo.TransactionLineage)
		*lineage = *expectedResponse
	})

	lineage, resp, err := svc.GetLineage(transactionID)

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, lineage)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_GetLineage_CorrectEndpoint(t *testing.T) {
	mockClient, svc := setupTransactionService()
	transactionID := "txn_8d2ce2f0-0d75-4a91-9d43-2ad2c2e6b9ad"
	endpoint := fmt.Sprintf("transactions/%s/lineage", transactionID)

	mockClient.On("NewRequest", endpoint, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil)

	_, _, _ = svc.GetLineage(transactionID)

	mockClient.AssertCalled(t, "NewRequest", endpoint, http.MethodGet, nil)
	mockClient.AssertExpectations(t)
}

func TestTransactionService_GetLineage_EmptyID(t *testing.T) {
	mockClient, svc := setupTransactionService()

	lineage, resp, err := svc.GetLineage("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transactionID is required")
	assert.Nil(t, lineage)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
	mockClient.AssertNotCalled(t, "CallWithRetry")
	mockClient.AssertExpectations(t)
}

func TestTransactionService_GetLineage_NewRequestError(t *testing.T) {
	mockClient, svc := setupTransactionService()
	transactionID := "txn_8d2ce2f0-0d75-4a91-9d43-2ad2c2e6b9ad"
	endpoint := fmt.Sprintf("transactions/%s/lineage", transactionID)

	mockClient.On("NewRequest", endpoint, http.MethodGet, nil).Return(nil, fmt.Errorf("request error"))

	lineage, resp, err := svc.GetLineage(transactionID)

	assert.Error(t, err)
	assert.Nil(t, lineage)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "CallWithRetry")
	mockClient.AssertExpectations(t)
}

func TestTransactionService_GetLineage_ServerError(t *testing.T) {
	mockClient, svc := setupTransactionService()
	transactionID := "txn_8d2ce2f0-0d75-4a91-9d43-2ad2c2e6b9ad"
	endpoint := fmt.Sprintf("transactions/%s/lineage", transactionID)
	expectedResp := &http.Response{StatusCode: http.StatusInternalServerError}

	mockClient.On("NewRequest", endpoint, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(expectedResp, fmt.Errorf("server error"))

	lineage, resp, err := svc.GetLineage(transactionID)

	assert.Error(t, err)
	assert.Nil(t, lineage)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}
