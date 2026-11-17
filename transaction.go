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
	// Deprecated: Core has no narration field on a distribution and ignores it.
	Narration string `json:"narration,omitempty"`
}

type ParentTransaction struct {
	Amount       float64  `json:"amount"`
	Reference    string   `json:"reference"`
	Precision    int64    `json:"precision"`
	Description  string   `json:"description"`
	Currency     string   `json:"currency"`
	Sources      []Source `json:"sources,omitempty"`
	Destinations []Source `json:"destinations,omitempty"`
	// Deprecated: Core 0.15.0 removed the rate column. Core ignores it on
	// create and never returns it.
	Rate          float64              `json:"rate,omitempty"`
	Source        string               `json:"source,omitempty"`
	Destination   string               `json:"destination,omitempty"`
	PreciseAmount *big.Int             `json:"precise_amount,omitempty"`
	SkipQueue     bool                 `json:"skip_queue"`
	Atomic        bool                 `json:"atomic,omitempty"`
	Status        PryTransactionStatus `json:"status"`
	MetaData      MetaData             `json:"meta_data,omitempty"`
	EffectiveDate *time.Time           `json:"effective_date"`
}

type CreateTransactionRequest struct {
	ParentTransaction
	Inflight           bool       `json:"inflight,omitempty"`
	InflightExpiryDate *time.Time `json:"inflight_expiry_date,omitempty"`
	InflightCommitDate *time.Time `json:"inflight_commit_date,omitempty"`
	ScheduledFor       *time.Time `json:"scheduled_for,omitempty"`
	AllowOverdraft     bool       `json:"allow_overdraft,omitempty"`
	// OverdraftLimit caps how far below zero the source may go when
	// AllowOverdraft is true, in major units (e.g. 500.00). Zero means no cap.
	OverdraftLimit float64 `json:"overdraft_limit,omitempty"`
	// DryRun previews the post without writing (Core 0.15.3+). Use CreateDryRun
	// so the response is typed as TransactionPreview.
	DryRun bool `json:"dry_run,omitempty"`
}

type Transaction struct {
	ParentTransaction
	CreatedAt           time.Time `json:"created_at"`
	TransactionID       string    `json:"transaction_id"`
	ParentTransactionID string    `json:"parent_transaction,omitempty"`
	// Queued is true when Core accepted the request onto the queue and has
	// not applied it yet.
	Queued bool `json:"queued,omitempty"`
	// Hash is the transaction's hash-chain digest (Core 0.15.0+).
	Hash     string `json:"hash,omitempty"`
	Inflight bool   `json:"inflight,omitempty"`
	// AmountString, AllowOverdraft and OverdraftLimit are echoed on the
	// create response only. Core does not store them as columns, so
	// Get and List return them empty (allow_overdraft is kept in MetaData).
	AmountString   string  `json:"amount_string,omitempty"`
	AllowOverdraft bool    `json:"allow_overdraft,omitempty"`
	OverdraftLimit float64 `json:"overdraft_limit,omitempty"`
	// Core sends the zero time (0001-01-01T00:00:00Z) for a date that was
	// never set; check IsZero.
	ScheduledFor       *time.Time `json:"scheduled_for,omitempty"`
	InflightExpiryDate *time.Time `json:"inflight_expiry_date,omitempty"`
	InflightCommitDate *time.Time `json:"inflight_commit_date,omitempty"`
}

type UpdateStatus struct {
	Status        InflightStatus `json:"status"`
	Amount        float64        `json:"amount"`
	PreciseAmount *big.Int       `json:"precise_amount"`
	SkipQueue     bool           `json:"skip_queue,omitempty"`
	// Reference is an optional idempotency key for this commit or void (Core
	// 0.15.7+). Repeating the same reference, action and amount returns the
	// original child; a different amount or action returns 409
	// TXN_DUPLICATE_REFERENCE.
	Reference string `json:"reference,omitempty"`
	// DryRun previews commit/void without settling the hold (Core 0.15.3+).
	// Use UpdateDryRun so the response is typed as TransactionPreview.
	DryRun bool `json:"dry_run,omitempty"`
}

type CreateBulkTransactionRequest struct {
	Transactions []CreateTransactionRequest `json:"transactions"`
	Inflight     bool                       `json:"inflight,omitempty"`
	Atomic       bool                       `json:"atomic,omitempty"`
	RunAsync     bool                       `json:"run_async,omitempty"`
	SkipQueue    bool                       `json:"skip_queue,omitempty"`
	// DryRun previews the batch without writing (Core 0.15.3+). run_async is
	// ignored on a bulk dry run. Use CreateBulkDryRun for the preview type.
	DryRun bool `json:"dry_run,omitempty"`
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
	// Description is the refund narration. When omitted or empty, Core copies
	// the original transaction's description.
	Description string `json:"description,omitempty"`
	// MetaData is merged onto metadata inherited from the original. Sent keys
	// replace matching keys; other keys stay. An empty object is ignored.
	MetaData MetaData `json:"meta_data,omitempty"`
	// DryRun previews the refund without writing (Core 0.15.3+). Use RefundDryRun
	// so the response is typed as TransactionPreview.
	DryRun bool `json:"dry_run,omitempty"`
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
	// DryRun previews the batch without committing (Core 0.15.3+). Use
	// BulkCommitInflightDryRun so the response is typed as BulkTransactionPreview.
	DryRun bool `json:"dry_run,omitempty"`
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
	// DryRun previews the batch without voiding (Core 0.15.3+). Use
	// BulkVoidInflightDryRun so the response is typed as BulkTransactionPreview.
	DryRun bool `json:"dry_run,omitempty"`
}

// PreviewRejection is why a dry-run projection would not apply.
type PreviewRejection struct {
	Code    string `json:"code"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

// BalanceProjection is one balance's current and projected state in a dry-run.
// Amounts are minor-unit strings.
type BalanceProjection struct {
	BalanceID                      string `json:"balance_id"`
	Role                           string `json:"role"`
	Currency                       string `json:"currency"`
	Virtual                        bool   `json:"virtual,omitempty"`
	CurrentBalance                 string `json:"current_balance"`
	CurrentAvailable               string `json:"current_available"`
	CurrentCreditBalance           string `json:"current_credit_balance"`
	CurrentDebitBalance            string `json:"current_debit_balance"`
	CurrentInflightDebitBalance    string `json:"current_inflight_debit_balance"`
	CurrentInflightCreditBalance   string `json:"current_inflight_credit_balance"`
	ResultingBalance               string `json:"resulting_balance"`
	ResultingAvailable             string `json:"resulting_available"`
	ResultingCreditBalance         string `json:"resulting_credit_balance"`
	ResultingDebitBalance          string `json:"resulting_debit_balance"`
	ResultingInflightDebitBalance  string `json:"resulting_inflight_debit_balance"`
	ResultingInflightCreditBalance string `json:"resulting_inflight_credit_balance"`
}

// LegProjection is one split leg in a multi-source or multi-destination dry-run.
type LegProjection struct {
	Identifier    string  `json:"identifier"`
	Role          string  `json:"role"`
	PreciseAmount string  `json:"precise_amount"`
	Amount        float64 `json:"amount"`
}

// TransactionPreview is the HTTP 200 body for a dry-run create, refund, or
// inflight update. It is not a recorded transaction.
type TransactionPreview struct {
	DryRun        bool              `json:"dry_run"`
	WouldApply    bool              `json:"would_apply"`
	Rejection     *PreviewRejection `json:"rejection,omitempty"`
	Operation     string            `json:"operation,omitempty"`
	Status        string            `json:"status,omitempty"`
	Reference     string            `json:"reference,omitempty"`
	Currency      string            `json:"currency"`
	Amount        float64           `json:"amount"`
	PreciseAmount string            `json:"precise_amount"`
	// Precision matches Core's preview payload (number). Existing create
	// requests keep ParentTransaction.Precision as int64.
	Precision float64             `json:"precision"`
	Balances  []BalanceProjection `json:"balances"`
	Legs      []LegProjection     `json:"legs,omitempty"`
	Notes     []string            `json:"notes,omitempty"`
}

// BulkTransactionPreview is the HTTP 200 body for a dry-run bulk create,
// bulk commit, or bulk void.
type BulkTransactionPreview struct {
	DryRun     bool                 `json:"dry_run"`
	WouldApply bool                 `json:"would_apply"`
	Cumulative bool                 `json:"cumulative"`
	Atomic     bool                 `json:"atomic"`
	Results    []TransactionPreview `json:"results"`
	Balances   []BalanceProjection  `json:"balances,omitempty"`
	Notes      []string             `json:"notes,omitempty"`
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
	if body.DryRun {
		return nil, nil, errUseDryRunMethod("CreateDryRun")
	}
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

// CreateDryRun previews POST /transactions without writing a transaction,
// queue entry, webhook, hook, or @ balance. The reference is not consumed.
// Core always returns HTTP 200; a projected rejection is WouldApply=false.
func (s *TransactionService) CreateDryRun(body CreateTransactionRequest) (*TransactionPreview, *http.Response, error) {
	body.DryRun = true
	if err := ValidateCreateTransacation(body); err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("transactions", http.MethodPost, body)
	if err != nil {
		return nil, nil, err
	}

	preview := new(TransactionPreview)
	resp, err := s.client.CallWithRetry(req, preview)
	if err != nil {
		return nil, resp, err
	}

	return preview, resp, nil
}

func (s *TransactionService) BulkCommitInflight(body BulkCommitInflightRequest) (*BulkCommitInflightResponse, *http.Response, error) {
	if body.DryRun {
		return nil, nil, errUseDryRunMethod("BulkCommitInflightDryRun")
	}
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

// BulkCommitInflightDryRun previews a bulk commit without settling holds.
func (s *TransactionService) BulkCommitInflightDryRun(body BulkCommitInflightRequest) (*BulkTransactionPreview, *http.Response, error) {
	body.DryRun = true
	if err := ValidateBulkCommitInflight(body); err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("transactions/inflight/bulk/commit", http.MethodPost, body)
	if err != nil {
		return nil, nil, err
	}

	preview := new(BulkTransactionPreview)
	resp, err := s.client.CallWithRetry(req, preview)
	if err != nil {
		return nil, resp, err
	}

	return preview, resp, nil
}

func (s *TransactionService) BulkVoidInflight(body BulkVoidInflightRequest) (*BulkVoidInflightResponse, *http.Response, error) {
	if body.DryRun {
		return nil, nil, errUseDryRunMethod("BulkVoidInflightDryRun")
	}
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

// BulkVoidInflightDryRun previews a bulk void without releasing holds.
func (s *TransactionService) BulkVoidInflightDryRun(body BulkVoidInflightRequest) (*BulkTransactionPreview, *http.Response, error) {
	body.DryRun = true
	if err := ValidateBulkVoidInflight(body); err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("transactions/inflight/bulk/void", http.MethodPost, body)
	if err != nil {
		return nil, nil, err
	}

	preview := new(BulkTransactionPreview)
	resp, err := s.client.CallWithRetry(req, preview)
	if err != nil {
		return nil, resp, err
	}

	return preview, resp, nil
}

func (s *TransactionService) CreateBulk(body CreateBulkTransactionRequest) (*CreateBulkTransactionResponse, *http.Response, error) {
	if body.DryRun {
		return nil, nil, errUseDryRunMethod("CreateBulkDryRun")
	}
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

// CreateBulkDryRun previews POST /transactions/bulk without writing.
// skip_queue still selects cumulative vs independent projection.
func (s *TransactionService) CreateBulkDryRun(body CreateBulkTransactionRequest) (*BulkTransactionPreview, *http.Response, error) {
	body.DryRun = true
	if err := ValidateCreateBulkTransaction(body); err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("transactions/bulk", http.MethodPost, body)
	if err != nil {
		return nil, nil, err
	}

	preview := new(BulkTransactionPreview)
	resp, err := s.client.CallWithRetry(req, preview)
	if err != nil {
		return nil, resp, err
	}

	return preview, resp, nil
}

func (s *TransactionService) Update(transactionID string, body UpdateStatus) (*Transaction, *http.Response, error) {
	if body.DryRun {
		return nil, nil, errUseDryRunMethod("UpdateDryRun")
	}
	//if transactionId is an empty string, return an error
	if transactionID == "" {
		return nil, nil, fmt.Errorf("transactionID is required")
	}
	u, err := endpointPath("transactions/inflight/%s", transactionID)
	if err != nil {
		return nil, nil, err
	}
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

// UpdateDryRun previews an inflight commit or void without settling the hold.
func (s *TransactionService) UpdateDryRun(transactionID string, body UpdateStatus) (*TransactionPreview, *http.Response, error) {
	if transactionID == "" {
		return nil, nil, fmt.Errorf("transactionID is required")
	}
	body.DryRun = true

	u, err := endpointPath("transactions/inflight/%s", transactionID)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(u, http.MethodPut, body)
	if err != nil {
		return nil, nil, err
	}

	preview := new(TransactionPreview)
	resp, err := s.client.CallWithRetry(req, preview)
	if err != nil {
		return nil, resp, err
	}

	return preview, resp, nil
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
		if body[0].DryRun {
			return nil, nil, errUseDryRunMethod("RefundDryRun")
		}
		if err := ValidateRefundTransaction(*body[0]); err != nil {
			return nil, nil, err
		}
		reqBody = body[0]
	}

	u, err := endpointPath("refund-transaction/%s", transactionID)
	if err != nil {
		return nil, nil, err
	}
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

// RefundDryRun previews a refund without writing a reversal.
func (s *TransactionService) RefundDryRun(transactionID string, body ...*RefundTransactionRequest) (*TransactionPreview, *http.Response, error) {
	if transactionID == "" {
		return nil, nil, fmt.Errorf("transactionID is required")
	}
	if len(body) > 1 {
		return nil, nil, fmt.Errorf("RefundDryRun accepts at most one optional request body")
	}

	reqBody := &RefundTransactionRequest{DryRun: true}
	if len(body) > 0 && body[0] != nil {
		copied := *body[0]
		copied.DryRun = true
		if err := ValidateRefundTransaction(copied); err != nil {
			return nil, nil, err
		}
		reqBody = &copied
	}

	u, err := endpointPath("refund-transaction/%s", transactionID)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(u, http.MethodPost, reqBody)
	if err != nil {
		return nil, nil, err
	}

	preview := new(TransactionPreview)
	resp, err := s.client.CallWithRetry(req, preview)
	if err != nil {
		return nil, resp, err
	}

	return preview, resp, nil
}

func (s *TransactionService) Get(transactionID string) (*Transaction, *http.Response, error) {
	if transactionID == "" {
		return nil, nil, fmt.Errorf("transactionID is required")
	}

	u, err := endpointPath("transactions/%s", transactionID)
	if err != nil {
		return nil, nil, err
	}
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

// GetByReference fetches a transaction by its reference. References are
// free-form, so "/", "%", spaces and Unicode are percent-escaped rather than
// rejected. A reference that is "." or "..", or whose "/"-separated parts
// include "." or ".." (for example "a/../b"), returns an error without
// sending a request, because a server or proxy that decodes %2F would read it
// as a dot segment.
func (s *TransactionService) GetByReference(reference string) (*Transaction, *http.Response, error) {
	if reference == "" {
		return nil, nil, fmt.Errorf("reference is required")
	}

	// References are free-form, so "/" and "%" are escaped rather than
	// rejected. Dot segments still are, since escaping leaves them as-is.
	if reference == "." || reference == ".." {
		return nil, nil, fmt.Errorf("invalid reference %q", reference)
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

// List returns transactions via GET /transactions, using Core's default page
// (limit=20, offset=0). Use ListWithOptions to page through results.
func (s *TransactionService) List() ([]Transaction, *http.Response, error) {
	return s.ListWithOptions(nil)
}

// ListWithOptions returns transactions via GET /transactions with optional
// pagination. A nil opts, or zero Limit and Offset, uses Core's default page.
func (s *TransactionService) ListWithOptions(opts *ListOptions) ([]Transaction, *http.Response, error) {
	q, err := listQuery(opts)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest("transactions", http.MethodGet, q)
	if err != nil {
		return nil, nil, err
	}

	var transactions []Transaction
	resp, err := s.client.CallWithRetry(req, &transactions)
	if err != nil {
		return nil, resp, err
	}

	return transactions, resp, nil
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

func errUseDryRunMethod(method string) error {
	return fmt.Errorf("dry_run is true: use Transaction.%s to receive a typed preview", method)
}
