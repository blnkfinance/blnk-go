package blnkgo

import (
	"fmt"
	"net/http"
	"time"
)

type ReconciliationService service

type ReconciliationUploadResp struct {
	UploadID    string `json:"upload_id"`
	RecordCount int    `json:"record_count"`
	Source      string `json:"source"`
}

// Criteria represents the filtering criteria.
type Criteria struct {
	Field          CriteriaField          `json:"field"`
	Operator       ReconciliationOperator `json:"operator"`
	AllowableDrift float64                `json:"allowable_drift,omitempty"` // Optional field
}

// Matcher represents a matching rule with multiple criteria.
type Matcher struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Criteria    []Criteria `json:"criteria"`
}

// RunReconData represents the data required to run a reconciliation process.
type RunReconData struct {
	UploadID         string                 `json:"upload_id"`
	Strategy         ReconciliationStrategy `json:"strategy"`
	DryRun           bool                   `json:"dry_run"`
	GroupingCriteria CriteriaField          `json:"grouping_criteria"`
	MatchingRuleIDs  []string               `json:"matching_rule_ids"`
}

type RunReconResp struct {
	Matcher
	RuleID    string `json:"rule_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ExternalTransaction is a single row of external data for instant reconciliation.
// Description, Date, and Source are optional at the Core HTTP layer; omit them when unset.
type ExternalTransaction struct {
	ID          string    `json:"id"`
	Amount      float64   `json:"amount"`
	Reference   string    `json:"reference"`
	Currency    string    `json:"currency"`
	Description string     `json:"description,omitempty"`
	Date        *time.Time `json:"date,omitempty"`
	Source      string     `json:"source,omitempty"`
}

// RunInstantReconData is the request body for POST /reconciliation/start-instant.
type RunInstantReconData struct {
	ExternalTransactions []ExternalTransaction  `json:"external_transactions"`
	Strategy             ReconciliationStrategy `json:"strategy"`
	GroupingCriteria     CriteriaField          `json:"grouping_criteria,omitempty"`
	DryRun               bool                   `json:"dry_run,omitempty"`
	MatchingRuleIDs      []string               `json:"matching_rule_ids"`
}

// RunInstantReconResp is returned when instant reconciliation is started.
type RunInstantReconResp struct {
	ReconciliationID string `json:"reconciliation_id"`
}

// Reconciliation is the status and counts for a reconciliation run.
type Reconciliation struct {
	ReconciliationID      string     `json:"reconciliation_id"`
	UploadID              string     `json:"upload_id"`
	Status                string     `json:"status"`
	MatchedTransactions   int        `json:"matched_transactions"`
	UnmatchedTransactions int        `json:"unmatched_transactions"`
	IsDryRun              bool       `json:"is_dry_run"`
	StartedAt             time.Time  `json:"started_at"`
	CompletedAt           *time.Time `json:"completed_at"`
}

const MaxInstantReconciliationItems = 10000

func (s *ReconciliationService) CreateMatchingRule(matcher Matcher) (*RunReconResp, *http.Response, error) {
	req, err := s.client.NewRequest("reconciliation/matching-rules", http.MethodPost, matcher)
	if err != nil {
		return nil, nil, err
	}

	reconResp := new(RunReconResp)
	resp, err := s.client.CallWithRetry(req, reconResp)
	if err != nil {
		return nil, resp, err
	}

	return reconResp, resp, nil
}

func (s *ReconciliationService) RunInstant(data RunInstantReconData) (*RunInstantReconResp, *http.Response, error) {
	if err := ValidateRunInstantReconData(data); err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("reconciliation/start-instant", http.MethodPost, data)
	if err != nil {
		return nil, nil, err
	}

	reconResp := new(RunInstantReconResp)
	resp, err := s.client.CallWithRetry(req, reconResp)
	if err != nil {
		return nil, resp, err
	}

	return reconResp, resp, nil
}

func (s *ReconciliationService) Get(reconciliationID string) (*Reconciliation, *http.Response, error) {
	if reconciliationID == "" {
		return nil, nil, fmt.Errorf("reconciliation id is required")
	}

	req, err := s.client.NewRequest(fmt.Sprintf("reconciliation/%s", reconciliationID), http.MethodGet, nil)
	if err != nil {
		return nil, nil, err
	}

	recon := new(Reconciliation)
	resp, err := s.client.CallWithRetry(req, recon)
	if err != nil {
		return nil, resp, err
	}

	return recon, resp, nil
}

func (s *ReconciliationService) Run(data RunReconData) (*RunReconResp, *http.Response, error) {
	req, err := s.client.NewRequest("reconciliation/start", http.MethodPost, data)
	if err != nil {
		return nil, nil, err
	}
	reconResp := new(RunReconResp)
	resp, err := s.client.CallWithRetry(req, reconResp)
	if err != nil {
		return nil, resp, err
	}

	return reconResp, resp, nil
}

func (s *ReconciliationService) Upload(source string, file interface{}, fileName string) (*ReconciliationUploadResp, *http.Response, error) {
	req, err := s.client.NewFileUploadRequest("reconciliation/upload", "file", file, fileName, map[string]string{
		"source": source,
	})

	if err != nil {
		return nil, nil, err
	}

	reconResp := new(ReconciliationUploadResp)
	resp, err := s.client.CallWithRetry(req, reconResp)
	if err != nil {
		return nil, resp, err
	}

	return reconResp, resp, nil
}

func NewReconciliationService(c ClientInterface) *ReconciliationService {
	return &ReconciliationService{client: c}
}
