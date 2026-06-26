package blnkgo_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransactionLineage_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		payload string
	}{
		{
			name: "docs example with empty shadow transactions",
			payload: `{
				"transaction_id": "txn_8d2ce2f0-0d75-4a91-9d43-2ad2c2e6b9ad",
				"shadow_transactions": []
			}`,
		},
		{
			name: "null shadow transactions",
			payload: `{
				"transaction_id": "txn_8d2ce2f0-0d75-4a91-9d43-2ad2c2e6b9ad",
				"shadow_transactions": null
			}`,
		},
		{
			name: "fund allocation with shadow transactions",
			payload: `{
				"transaction_id": "txn_main_123",
				"fund_allocation": [{"provider": "stripe", "amount": "1000"}],
				"shadow_transactions": [{
					"transaction_id": "txn_shadow_1",
					"reference": "shadow-ref",
					"status": "APPLIED",
					"precise_amount": "1000"
				}]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lineage blnkgo.TransactionLineage
			err := json.Unmarshal([]byte(tt.payload), &lineage)
			assert.NoError(t, err)
			assert.NotEmpty(t, lineage.TransactionID)
			if tt.name == "fund allocation with shadow transactions" {
				assert.Len(t, lineage.FundAllocation, 1)
				assert.Len(t, lineage.ShadowTransactions, 1)
				assert.Equal(t, "txn_shadow_1", lineage.ShadowTransactions[0]["transaction_id"])
			}
		})
	}
}

func TestTransactionService_GetLineage_Success(t *testing.T) {
	mockClient, svc := setupTransactionService()
	transactionID := "txn_8d2ce2f0-0d75-4a91-9d43-2ad2c2e6b9ad"

	expectedResponse := &blnkgo.TransactionLineage{
		TransactionID: transactionID,
		FundAllocation: []map[string]interface{}{
			{"provider": "stripe", "amount": "1000"},
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

func TestTransactionService_GetLineage_ServerError(t *testing.T) {
	mockClient, svc := setupTransactionService()
	transactionID := "txn_nonexistent"
	endpoint := fmt.Sprintf("transactions/%s/lineage", transactionID)
	expectedResp := &http.Response{StatusCode: http.StatusNotFound}

	mockClient.On("NewRequest", endpoint, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(expectedResp, fmt.Errorf("not found"))

	lineage, resp, err := svc.GetLineage(transactionID)

	assert.Error(t, err)
	assert.Nil(t, lineage)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}
