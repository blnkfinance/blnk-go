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

func TestLedgerBalanceService_UpdateIdentity_Success(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	balanceID := "bln_5ce86029-3c2e-4e2a-aae2-7fb931ca4c4f"
	body := blnkgo.UpdateBalanceIdentityRequest{
		IdentityID: "idt_3b63c8da-af29-4cc3-ad38-df17d87456e6",
	}

	expectedResponse := &blnkgo.UpdateBalanceIdentityResponse{
		Message: "Balance identity updated successfully",
	}

	endpoint := fmt.Sprintf("balances/%s/identity", balanceID)
	mockClient.On("NewRequest", endpoint, http.MethodPut, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		response := args.Get(1).(*blnkgo.UpdateBalanceIdentityResponse)
		*response = *expectedResponse
	})

	response, resp, err := svc.UpdateIdentity(balanceID, body)

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_UpdateIdentity_CorrectEndpoint(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	balanceID := "bln_5ce86029-3c2e-4e2a-aae2-7fb931ca4c4f"
	body := blnkgo.UpdateBalanceIdentityRequest{IdentityID: "idt_3b63c8da-af29-4cc3-ad38-df17d87456e6"}
	endpoint := fmt.Sprintf("balances/%s/identity", balanceID)

	mockClient.On("NewRequest", endpoint, http.MethodPut, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil)

	_, _, _ = svc.UpdateIdentity(balanceID, body)

	mockClient.AssertCalled(t, "NewRequest", endpoint, http.MethodPut, body)
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_UpdateIdentity_EmptyBalanceID(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	body := blnkgo.UpdateBalanceIdentityRequest{IdentityID: "idt_3b63c8da-af29-4cc3-ad38-df17d87456e6"}

	response, resp, err := svc.UpdateIdentity("", body)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "balanceID is required")
	assert.Nil(t, response)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
	mockClient.AssertNotCalled(t, "CallWithRetry")
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_UpdateIdentity_EmptyIdentityID(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	body := blnkgo.UpdateBalanceIdentityRequest{}

	response, resp, err := svc.UpdateIdentity("bln_5ce86029-3c2e-4e2a-aae2-7fb931ca4c4f", body)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "identity_id is required")
	assert.Nil(t, response)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "NewRequest")
	mockClient.AssertNotCalled(t, "CallWithRetry")
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_UpdateIdentity_NewRequestError(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	balanceID := "bln_5ce86029-3c2e-4e2a-aae2-7fb931ca4c4f"
	body := blnkgo.UpdateBalanceIdentityRequest{IdentityID: "idt_3b63c8da-af29-4cc3-ad38-df17d87456e6"}
	endpoint := fmt.Sprintf("balances/%s/identity", balanceID)

	mockClient.On("NewRequest", endpoint, http.MethodPut, body).Return(nil, fmt.Errorf("request error"))

	response, resp, err := svc.UpdateIdentity(balanceID, body)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Nil(t, resp)
	mockClient.AssertNotCalled(t, "CallWithRetry")
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_UpdateIdentity_ServerError(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	balanceID := "bln_5ce86029-3c2e-4e2a-aae2-7fb931ca4c4f"
	body := blnkgo.UpdateBalanceIdentityRequest{IdentityID: "idt_3b63c8da-af29-4cc3-ad38-df17d87456e6"}
	endpoint := fmt.Sprintf("balances/%s/identity", balanceID)
	expectedResp := &http.Response{StatusCode: http.StatusInternalServerError}

	mockClient.On("NewRequest", endpoint, http.MethodPut, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(expectedResp, fmt.Errorf("server error"))

	response, resp, err := svc.UpdateIdentity(balanceID, body)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestUpdateBalanceIdentityResponse_UnmarshalJSON(t *testing.T) {
	payload := `{"message":"Balance identity updated successfully"}`

	var response blnkgo.UpdateBalanceIdentityResponse
	err := json.Unmarshal([]byte(payload), &response)

	assert.NoError(t, err)
	assert.Equal(t, "Balance identity updated successfully", response.Message)
}
