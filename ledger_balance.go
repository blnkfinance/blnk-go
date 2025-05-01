package blnkgo

import (
	"fmt"
	"math/big"
	"net/http"
	"time"
)

type LedgerBalanceService service

type LedgerBalance struct {
	BalanceID             string                 `json:"balance_id"`
	Balance               *big.Int               `json:"balance"`
	Version               int64                  `json:"version"`
	InflightBalance       *big.Int               `json:"inflight_balance"`
	CreditBalance         *big.Int               `json:"credit_balance"`
	InflightCreditBalance *big.Int               `json:"inflight_credit_balance"`
	DebitBalance          *big.Int               `json:"debit_balance"`
	InflightDebitBalance  *big.Int               `json:"inflight_debit_balance"`
	QueuedDebitBalance    *big.Int               `json:"queued_debit_balance,omitempty"`
	QueuedCreditBalance   *big.Int               `json:"queued_credit_balance,omitempty"`
	CurrencyMultiplier    float64                `json:"currency_multiplier"`
	Precision             int                    `json:"precision"`
	LedgerID              string                 `json:"ledger_id"`
	IdentityID            string                 `json:"identity_id"`
	Indicator             string                 `json:"indicator"`
	Currency              string                 `json:"currency"`
	CreatedAt             time.Time              `json:"created_at"`
	InflightExpiresAt     time.Time              `json:"inflight_expires_at"`
	MetaData              map[string]interface{} `json:"meta_data,omitempty"`
}

type CreateLedgerBalanceRequest struct {
	LedgerID   string                 `json:"ledger_id"`
	IdentityID string                 `json:"identity_id,omitempty"`
	Currency   string                 `json:"currency"`
	MetaData   map[string]interface{} `json:"meta_data,omitempty"`
}

func (s *LedgerBalanceService) Create(body CreateLedgerBalanceRequest) (*LedgerBalance, *http.Response, error) {
	req, err := s.client.NewRequest("balances", http.MethodPost, body)
	if err != nil {
		return nil, nil, err
	}

	ledgerBalance := new(LedgerBalance)
	resp, err := s.client.CallWithRetry(req, ledgerBalance)
	if err != nil {
		return nil, resp, err
	}

	return ledgerBalance, resp, nil
}

func (s *LedgerBalanceService) Get(balanceID string) (*LedgerBalance, *http.Response, error) {
	if balanceID == "" {
		return nil, nil, fmt.Errorf("invalid: id is required")
	}
	u := fmt.Sprintf("balances/%s", balanceID)
	req, err := s.client.NewRequest(u, http.MethodGet, nil)
	if err != nil {
		return nil, nil, err
	}
	ledgerBalance := new(LedgerBalance)
	resp, err := s.client.CallWithRetry(req, ledgerBalance)
	if err != nil {
		return nil, resp, err
	}
	return ledgerBalance, resp, nil
}

func (s *LedgerBalanceService) GetByIndicator(indicator string, currency string) (*LedgerBalance, *http.Response, error) {
	if indicator == "" {
		return nil, nil, fmt.Errorf("indicator is required")
	}
	if currency == "" {
		return nil, nil, fmt.Errorf("currency is required")
	}
	u := fmt.Sprintf("balances/indicator/%s/currency/%s", indicator, currency)
	req, err := s.client.NewRequest(u, http.MethodGet, nil)
	if err != nil {
		return nil, nil, err
	}
	ledgerBalance := new(LedgerBalance)
	resp, err := s.client.CallWithRetry(req, ledgerBalance)
	if err != nil {
		return nil, resp, err
	}
	return ledgerBalance, resp, nil
}

func NewLedgerBalanceService(c ClientInterface) *LedgerBalanceService {
	return &LedgerBalanceService{client: c}
}
