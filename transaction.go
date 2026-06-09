package blnkgo

import (
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"time"
)

type TransactionService service

type Source struct {
	Identifier   string       `json:"identifier"`
	Distribution Distribution `json:"distribution"`
	Narration    string       `json:"narration,omitempty"`
}

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
	Atomic        bool                   `json:"atomic,omitempty"`
	Status        PryTransactionStatus   `json:"status"`
	MetaData      map[string]interface{} `json:"meta_data,omitempty"`
	EffectiveDate *time.Time             `json:"effective_date"`
}

type CreateTransactionRequest struct {
	ParentTransaction
	Inflight           bool       `json:"inflight,omitempty"`
	InflightExpiryDate *time.Time `json:"inflight_expiry_date,omitempty"`
	InflightCommitDate *time.Time `json:"inflight_commit_date,omitempty"`
	ScheduledFor       *time.Time `json:"scheduled_for,omitempty"`
	AllowOverdraft     bool       `json:"allow_overdraft,omitempty"`
}

type Transaction struct {
	ParentTransaction
	CreatedAt     time.Time `json:"created_at"`
	TransactionID string    `json:"transaction_id"`
}

type UpdateStatus struct {
	Status        InflightStatus `json:"status"`
	Amount        float64        `json:"amount"`
	PreciseAmount *big.Int       `json:"precise_amount"`
}

type CreateBulkTransactionRequest struct {
	Transactions []CreateTransactionRequest `json:"transactions"`
	Inflight     bool                       `json:"inflight,omitempty"`
	Atomic       bool                       `json:"atomic,omitempty"`
	RunAsync     bool                       `json:"run_async,omitempty"`
	SkipQueue    bool                       `json:"skip_queue,omitempty"`
}

type CreateBulkTransactionResponse struct {
	BatchID          string `json:"batch_id"`
	Status           string `json:"status"`
	TransactionCount int    `json:"transaction_count,omitempty"`
	Message          string `json:"message,omitempty"`
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

func (s *TransactionService) CreateBulk(body CreateBulkTransactionRequest) (*CreateBulkTransactionResponse, *http.Response, error) {
	if err := ValidateCreateBulkTransaction(body); err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("transactions/bulk", http.MethodPost, body)
	if err != nil {
		return nil, nil, err
	}

	response := new(CreateBulkTransactionResponse)
	resp, err := s.client.CallWithRetry(req, response)
	if err != nil {
		return nil, resp, err
	}

	return response, resp, nil
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

func (s *TransactionService) GetByReference(reference string) (*Transaction, *http.Response, error) {
	if reference == "" {
		return nil, nil, fmt.Errorf("reference is required")
	}

	u := fmt.Sprintf("transactions/reference/%s", url.PathEscape(reference))
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

func (s *TransactionService) Filter(params FilterParams) (*FilterResponse, *http.Response, error) {
	req, err := s.client.NewRequest("transactions/filter", http.MethodPost, params)
	if err != nil {
		return nil, nil, err
	}

	var filterResponse FilterResponse
	resp, err := s.client.CallWithRetry(req, &filterResponse)
	if err != nil {
		return nil, resp, err
	}

	return &filterResponse, resp, nil
}

func NewTransactionService(client ClientInterface) *TransactionService {
	return &TransactionService{client: client}
}
