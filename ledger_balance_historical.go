package blnkgo

import (
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"time"
)

type BalanceDetails struct {
	Balance       *big.Int `json:"balance"`
	BalanceID     string   `json:"balance_id"`
	CreditBalance *big.Int `json:"credit_balance"`
	Currency      string   `json:"currency"`
	DebitBalance  *big.Int `json:"debit_balance"`
}

type LedgerBalanceHistorical struct {
	Balance    BalanceDetails `json:"balance"`
	FromSource bool           `json:"from_source"`
	Timestamp  time.Time      `json:"timestamp"`
}

func (s *LedgerBalanceService) GetHistorical(balanceID string, timestamp time.Time, fromSource bool) (*LedgerBalanceHistorical, *http.Response, error) {
	if balanceID == "" {
		return nil, nil, fmt.Errorf("invalid: balanceID is required")
	}
	if timestamp.IsZero() {
		return nil, nil, fmt.Errorf("invalid: timestamp is required")
	}
	// Use RFC3339 with offset (e.g., 2025-08-30T01:38:30-03:00)
	ts := timestamp.Format(time.RFC3339)
	u, err := endpointPath("balances/%s/at", balanceID)
	if err != nil {
		return nil, nil, err
	}
	u += "?timestamp=" + url.QueryEscape(ts)
	if fromSource {
		u += "&from_source=true"
	}
	req, err := s.client.NewRequest(u, http.MethodGet, nil)
	if err != nil {
		return nil, nil, err
	}

	responseData := new(LedgerBalanceHistorical)
	resp, err := s.client.CallWithRetry(req, responseData)
	if err != nil {
		return nil, resp, err
	}

	return responseData, resp, nil
}
