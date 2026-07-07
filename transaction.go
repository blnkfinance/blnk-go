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
	Identifier          string       `json:"identifier"`
	Distribution        Distribution `json:"distribution,omitempty"`
	PreciseDistribution string       `json:"precise_distribution,omitempty"`
	Narration           string       `json:"narration,omitempty"`
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
	MetaData      MetaData             `json:"meta_data,omitempty"`
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
	CreatedAt           time.Time `json:"created_at"`
	TransactionID       string    `json:"transaction_id"`
	ParentTransactionID string    `json:"parent_transaction,omitempty"`
	Queued              bool      `json:"queued,omitempty"`
}

type UpdateStatus struct {
	Status        InflightStatus `json:"status"`
	Amount        float64        `json:"amount"`
	PreciseAmount *big.Int       `json:"precise_amount"`
	SkipQueue     bool           `json:"skip_queue,omitempty"`
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

// RefundTransactionRequest is the optional body for POST /refund-transaction/{id}.
// Omit the body (pass nil) to queue the refund using Core defaults.
type RefundTransactionRequest struct {
	SkipQueue bool `json:"skip_queue,omitempty"`
}

// MaxBulkInflightItems caps the number of transactions accepted in a single
// bulk commit or bulk void call.
const MaxBulkInflightItems = 100

// MaxBulkCreateItems caps the number of transactions accepted in a single
// CreateBulk call (POST /transactions/bulk).
const MaxBulkCreateItems = 10000

// BulkCommitInflightItem describes one transaction in a bulk commit request.
// Zero amount means commit the full remaining inflight amount; non-zero performs
// a partial commit. PreciseAmount, when set, takes precedence over Amount.
type BulkCommitInflightItem struct {
	TransactionID string   `json:"transaction_id"`
	Amount        float64  `json:"amount,omitempty"`
	PreciseAmount *big.Int `json:"precise_amount,omitempty"`
}

// BulkCommitInflightRequest commits many independently-created inflight
// transactions in one call.
type BulkCommitInflightRequest struct {
	Transactions []BulkCommitInflightItem `json:"transactions"`
	SkipQueue    bool                     `json:"skip_queue,omitempty"`
}

// BulkCommitInflightResult is the per-item outcome in BulkCommitInflightResponse.
type BulkCommitInflightResult struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	Code          string `json:"code,omitempty"`
	Message       string `json:"message,omitempty"`
}

// BulkCommitInflightResponse is the envelope returned by bulk commit inflight.
type BulkCommitInflightResponse struct {
	Succeeded int                        `json:"succeeded"`
	Failed    int                        `json:"failed"`
	Results   []BulkCommitInflightResult `json:"results"`
}

// BulkVoidInflightRequest voids many independently-created inflight
// transactions in one call.
type BulkVoidInflightRequest struct {
	TransactionIDs []string `json:"transaction_ids"`
	SkipQueue      bool     `json:"skip_queue,omitempty"`
}

// BulkVoidInflightResult is the per-item outcome in BulkVoidInflightResponse.
type BulkVoidInflightResult struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	Code          string `json:"code,omitempty"`
	Message       string `json:"message,omitempty"`
}

// BulkVoidInflightResponse is the envelope returned by bulk void inflight.
type BulkVoidInflightResponse struct {
	Succeeded int                      `json:"succeeded"`
	Failed    int                      `json:"failed"`
	Results   []BulkVoidInflightResult `json:"results"`
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

func (s *TransactionService) BulkCommitInflight(body BulkCommitInflightRequest) (*BulkCommitInflightResponse, *http.Response, error) {
	if err := ValidateBulkCommitInflight(body); err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("transactions/inflight/bulk/commit", http.MethodPost, body)
	if err != nil {
		return nil, nil, err
	}

	response := new(BulkCommitInflightResponse)
	resp, err := s.client.CallWithRetry(req, response)
	if err != nil {
		return nil, resp, err
	}

	return response, resp, nil
}

func (s *TransactionService) BulkVoidInflight(body BulkVoidInflightRequest) (*BulkVoidInflightResponse, *http.Response, error) {
	if err := ValidateBulkVoidInflight(body); err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("transactions/inflight/bulk/void", http.MethodPost, body)
	if err != nil {
		return nil, nil, err
	}

	response := new(BulkVoidInflightResponse)
	resp, err := s.client.CallWithRetry(req, response)
	if err != nil {
		return nil, resp, err
	}

	return response, resp, nil
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

func (s *TransactionService) Refund(transactionID string, body ...*RefundTransactionRequest) (*Transaction, *http.Response, error) {
	if transactionID == "" {
		return nil, nil, fmt.Errorf("transactionID is required")
	}
	if len(body) > 1 {
		return nil, nil, fmt.Errorf("Refund accepts at most one optional request body")
	}

	var reqBody interface{}
	if len(body) > 0 && body[0] != nil {
		if err := ValidateRefundTransaction(*body[0]); err != nil {
			return nil, nil, err
		}
		reqBody = body[0]
	}

	u := fmt.Sprintf("refund-transaction/%s", transactionID)
	req, err := s.client.NewRequest(u, http.MethodPost, reqBody)
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
