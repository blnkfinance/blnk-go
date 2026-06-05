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

// TestBalanceLineage_UnmarshalJSON verifies BalanceLineage decodes the API
// response contract from docs.blnkfinance.com/reference/get-balance-lineage.
func TestBalanceLineage_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		payload string
	}{
		{
			name: "docs example with string amounts",
			payload: `{
				"balance_id": "bln_5ce86029-3c2e-4e2a-aae2-7fb931ca4c4f",
				"aggregate_balance_id": "bln_aggregate_shadow_balance_id",
				"total_with_lineage": "7500",
				"providers": [{
					"provider": "stripe",
					"amount": "10000",
					"available": "7500",
					"spent": "2500",
					"shadow_balance_id": "bln_shadow_balance_id"
				}]
			}`,
		},
		{
			name: "core serialization with numeric amounts",
			payload: `{
				"balance_id": "bln_5ce86029-3c2e-4e2a-aae2-7fb931ca4c4f",
				"aggregate_balance_id": "bln_aggregate_shadow_balance_id",
				"total_with_lineage": 7500,
				"providers": [{
					"provider": "stripe",
					"amount": 10000,
					"available": 7500,
					"spent": 2500,
					"shadow_balance_id": "bln_shadow_balance_id"
				}]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lineage blnkgo.BalanceLineage
			err := json.Unmarshal([]byte(tt.payload), &lineage)
			assert.NoError(t, err)

			assert.Equal(t, "bln_5ce86029-3c2e-4e2a-aae2-7fb931ca4c4f", lineage.BalanceID)
			assert.Equal(t, "bln_aggregate_shadow_balance_id", lineage.AggregateBalanceID)
			assert.Equal(t, 0, big.NewInt(7500).Cmp(lineage.TotalWithLineage))
			assert.Len(t, lineage.Providers, 1)

			provider := lineage.Providers[0]
			assert.Equal(t, "stripe", provider.Provider)
			assert.Equal(t, 0, big.NewInt(10000).Cmp(provider.Amount))
			assert.Equal(t, 0, big.NewInt(7500).Cmp(provider.Available))
			assert.Equal(t, 0, big.NewInt(2500).Cmp(provider.Spent))
			assert.Equal(t, "bln_shadow_balance_id", provider.ShadowBalanceID)
		})
	}
}

func TestLedgerBalanceService_GetLineage_Success(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	balanceID := "bln_5ce86029-3c2e-4e2a-aae2-7fb931ca4c4f"

	expectedResponse := &blnkgo.BalanceLineage{
		BalanceID:          balanceID,
		AggregateBalanceID: "bln_aggregate_shadow_balance_id",
		TotalWithLineage:   big.NewInt(7500),
		Providers: []blnkgo.LineageProviderBreakdown{
			{
				Provider:        "stripe",
				Amount:          big.NewInt(10000),
				Available:       big.NewInt(7500),
				Spent:           big.NewInt(2500),
				ShadowBalanceID: "bln_shadow_balance_id",
			},
		},
	}

	endpoint := fmt.Sprintf("balances/%s/lineage", balanceID)
	mockClient.On("NewRequest", endpoint, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		lineage := args.Get(1).(*blnkgo.BalanceLineage)
		*lineage = *expectedResponse
	})

	lineage, resp, err := svc.GetLineage(balanceID)

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, lineage)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_GetLineage_CorrectEndpoint(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	balanceID := "bln_5ce86029-3c2e-4e2a-aae2-7fb931ca4c4f"
	endpoint := fmt.Sprintf("balances/%s/lineage", balanceID)

	mockClient.On("NewRequest", endpoint, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil)

	_, _, _ = svc.GetLineage(balanceID)

	mockClient.AssertCalled(t, "NewRequest", endpoint, http.MethodGet, nil)
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_GetLineage_EmptyID(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()

	lineage, resp, err := svc.GetLineage("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "balanceID is required")
	assert.Nil(t, lineage)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
	mockClient.AssertNotCalled(t, "CallWithRetry")
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_GetLineage_NewRequestError(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	balanceID := "bln_5ce86029-3c2e-4e2a-aae2-7fb931ca4c4f"
	endpoint := fmt.Sprintf("balances/%s/lineage", balanceID)

	mockClient.On("NewRequest", endpoint, http.MethodGet, nil).Return(nil, fmt.Errorf("request error"))

	lineage, resp, err := svc.GetLineage(balanceID)

	assert.Error(t, err)
	assert.Nil(t, lineage)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "CallWithRetry")
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_GetLineage_ServerError(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	balanceID := "bln_5ce86029-3c2e-4e2a-aae2-7fb931ca4c4f"
	endpoint := fmt.Sprintf("balances/%s/lineage", balanceID)
	expectedResp := &http.Response{StatusCode: http.StatusInternalServerError}

	mockClient.On("NewRequest", endpoint, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(expectedResp, fmt.Errorf("server error"))

	lineage, resp, err := svc.GetLineage(balanceID)

	assert.Error(t, err)
	assert.Nil(t, lineage)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}
