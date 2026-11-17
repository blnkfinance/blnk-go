package blnkgo_test

import (
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLedgerBalanceService_GetHistorical_Success(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	balanceID := "bal-123"
	timestamp := time.Date(2023, 8, 29, 12, 0, 0, 0, time.FixedZone("-03:00", -3*60*60))
	endpoint := fmt.Sprintf("balances/%s/at?timestamp=%s", balanceID, url.QueryEscape(timestamp.Format(time.RFC3339)))
	mockClient.On("NewRequest", endpoint, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{}, nil).Run(func(args mock.Arguments) {
		respData := args.Get(1).(*blnkgo.LedgerBalanceHistorical)
		respData.Balance.Balance = big.NewInt(1000)
		respData.Balance.BalanceID = balanceID
		respData.Balance.CreditBalance = big.NewInt(2000)
		respData.Balance.Currency = "USD"
		respData.Balance.DebitBalance = big.NewInt(1000)
		respData.FromSource = true
		respData.Timestamp = timestamp
	})
	bal, resp, err := svc.GetHistorical(balanceID, timestamp, false)
	assert.NoError(t, err)
	assert.NotNil(t, bal)
	assert.Equal(t, big.NewInt(1000), bal.Balance.Balance)
	assert.Equal(t, balanceID, bal.Balance.BalanceID)
	assert.Equal(t, big.NewInt(2000), bal.Balance.CreditBalance)
	assert.Equal(t, "USD", bal.Balance.Currency)
	assert.Equal(t, big.NewInt(1000), bal.Balance.DebitBalance)
	assert.True(t, bal.FromSource)
	assert.True(t, bal.Timestamp.Equal(timestamp))
	assert.NotNil(t, resp)
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_GetHistorical_InvalidBalanceID(t *testing.T) {
	_, svc := setupLedgerBalanceService()
	bal, resp, err := svc.GetHistorical("", time.Now(), false)
	assert.Error(t, err)
	assert.Nil(t, bal)
	assert.Nil(t, resp)
}

func TestLedgerBalanceService_GetHistorical_InvalidTimestamp(t *testing.T) {
	_, svc := setupLedgerBalanceService()
	bal, resp, err := svc.GetHistorical("bal-123", time.Time{}, false)
	assert.Error(t, err)
	assert.Nil(t, bal)
	assert.Nil(t, resp)
}

func TestLedgerBalanceService_GetHistorical_NewRequestError(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	mockClient.On("NewRequest", mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("request error")) // Mantém compatível
	bal, resp, err := svc.GetHistorical("bal-123", time.Now(), false)
	assert.Error(t, err)
	assert.Nil(t, bal)
	assert.Nil(t, resp)
	mockClient.AssertExpectations(t)
}

func TestLedgerBalanceService_GetHistorical_CallWithRetryError(t *testing.T) {
	mockClient, svc := setupLedgerBalanceService()
	mockClient.On("NewRequest", mock.Anything, mock.Anything, mock.Anything).Return(&http.Request{}, nil) // Mantém compatível
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("call error"))
	bal, resp, err := svc.GetHistorical("bal-123", time.Now(), false)
	assert.Error(t, err)
	assert.Nil(t, bal)
	assert.Nil(t, resp)
	mockClient.AssertExpectations(t)
}
