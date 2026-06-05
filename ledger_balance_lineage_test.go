package blnkgo_test

import (
	"fmt"
	"math/big"
	"net/http"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
