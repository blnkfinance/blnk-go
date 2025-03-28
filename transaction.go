package blnkgo

import (
	"fmt"
	"math/big"
	"net/http"
	"time"
)

type TransactionService service

// MultipleSourcesT represents multiple sources for a transaction.
type Source struct {
	Identifier   string       `json:"identifier"`
	Distribution Distribution `json:"distribution"`
	Narration    string       `json:"narration,omitempty"`
}

// CreateTransactionResponse represents the response for creating a transaction.
type ParentTransaction struct {
	Amount        float64                `json:"amount"`
	Reference     string                 `json:"reference"`
	Precision     int64                  `json:"precision"`
	Description   string                 `json:"description"`
	Currency      string                 `json:"currency"`
	Sources       []Source               `json:"sources,omitempty"`
	Destinations  []Source               `json:"destinations,omitempty"`
	Rate          float64                `json:"rate,omitempty"`
	Source        string                 `json:"source,omitempty"`
	Destination   string                 `json:"destination,omitempty"`
	PreciseAmount *big.Int               `json:"precise_amount,omitempty"`
	SkipQueue     bool                   `json:"skip_queue"`
	Status        PryTransactionStatus   `json:"status"`
	MetaData      map[string]interface{} `json:"meta_data,omitempty"`
}

type CreateTransactionRequest struct {
	ParentTransaction
	Inflight           bool       `json:"inflight,omitempty"`
	InflightExpiryDate *time.Time `json:"inflight_expiry_date,omitempty"`
	ScheduledFor       *time.Time `json:"scheduled_for,omitempty"`
	AllowOverdraft     bool       `json:"allow_overdraft,omitempty"`
}

type Transaction struct {
	ParentTransaction
	CreatedAt     time.Time `json:"created_at"`
	TransactionID string    `json:"transaction_id"`
}

type UpdateStatus struct {
	Status InflightStatus `json:"status"`
	Amount float64        `json:"amount"`
}

func (s *TransactionService) Create(body CreateTransactionRequest) (*Transaction, *http.Response, error) {
	//validate the trannsaction
	if err := ValidateCreateTransacation(body); err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("transactions", http.MethodPost, body)
	if err != nil {
		return nil, nil, err
	}

	transaction := new(Transaction)
	resp, err := s.client.CallWithRetry(req, transaction)
	if err != nil {
		return nil, resp, err
	}

	return transaction, resp, nil
}
func (s *TransactionService) Update(transactionID string, body UpdateStatus) (*Transaction, *http.Response, error) {
	//if transactionId is an empty string, return an error
	if transactionID == "" {
		return nil, nil, fmt.Errorf("transactionID is required")
	}
	u := fmt.Sprintf("transactions/inflight/%s", transactionID)
	req, err := s.client.NewRequest(u, http.MethodPut, body)
	if err != nil {
		return nil, nil, err
	}

	transaction := new(Transaction)
	resp, err := s.client.CallWithRetry(req, transaction)
	if err != nil {
		return nil, resp, err
	}

	return transaction, resp, nil
}

func (s *TransactionService) Refund(transactionID string) (*Transaction, *http.Response, error) {
	u := fmt.Sprintf("refund-transaction/%s", transactionID)
	req, err := s.client.NewRequest(u, http.MethodPost, nil)
	if err != nil {
		return nil, nil, err
	}

	transaction := new(Transaction)
	resp, err := s.client.CallWithRetry(req, transaction)
	if err != nil {
		return nil, resp, err
	}

	return transaction, resp, nil
}

func (s *TransactionService) Get(transactionID string) (*Transaction, *http.Response, error) {
	if transactionID == "" {
		return nil, nil, fmt.Errorf("transactionID is required")
	}

	u := fmt.Sprintf("transactions/%s", transactionID)
	req, err := s.client.NewRequest(u, http.MethodGet, nil)
	if err != nil {
		return nil, nil, err
	}

	transaction := new(Transaction)
	resp, err := s.client.CallWithRetry(req, transaction)
	if err != nil {
		return nil, resp, err
	}

	return transaction, resp, nil
}

func NewTransactionService(client ClientInterface) *TransactionService {
	return &TransactionService{client: client}
}
