package blnkgo

import (
	"fmt"
	"net/http"
	"time"
)

type LedgerService service

type Ledger struct {
	LedgerID  string                 `json:"ledger_id"`
	Name      string                 `json:"name"`
	CreatedAt time.Time              `json:"created_at"`
	MetaData  MetaData  `json:"meta_data,omitempty"`
}

type CreateLedgerRequest struct {
	Name     string   `json:"name"`
	MetaData MetaData `json:"meta_data,omitempty"`
}

type UpdateLedgerRequest struct {
	Name string `json:"name"`
}

// List returns ledgers via GET /ledgers, using Core's default page
// (limit=10, offset=0). Use ListWithOptions to page through results.
func (s *LedgerService) List() ([]Ledger, *http.Response, error) {
	return s.ListWithOptions(nil)
}

// ListWithOptions returns ledgers via GET /ledgers with optional
// pagination. A nil opts, or zero Limit and Offset, uses Core's default page.
func (s *LedgerService) ListWithOptions(opts *ListOptions) ([]Ledger, *http.Response, error) {
	q, err := listQuery(opts)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest("ledgers", http.MethodGet, q)
	if err != nil {
		return nil, nil, err
	}
	var ledgers []Ledger
	resp, err := s.client.CallWithRetry(req, &ledgers)
	if err != nil {
		return nil, resp, err
	}
	return ledgers, resp, nil
}

func (s *LedgerService) Get(id string) (*Ledger, *http.Response, error) {
	if id == "" {
		return nil, nil, fmt.Errorf("invalid: id is required")
	}
	u, err := endpointPath("ledgers/%s", id)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(u, http.MethodGet, nil)
	if err != nil {
		return nil, nil, err
	}

	ledger := new(Ledger)
	resp, err := s.client.CallWithRetry(req, ledger)
	if err != nil {
		return nil, resp, err
	}

	return ledger, resp, nil
}

func (s *LedgerService) Create(body CreateLedgerRequest) (*Ledger, *http.Response, error) {
	req, err := s.client.NewRequest("ledgers", http.MethodPost, body)
	if err != nil {
		return nil, nil, err
	}

	ledger := new(Ledger)
	resp, err := s.client.CallWithRetry(req, ledger)
	if err != nil {
		return nil, resp, err
	}

	return ledger, resp, nil
}

func (s *LedgerService) Update(id string, body UpdateLedgerRequest) (*Ledger, *http.Response, error) {
	if id == "" {
		return nil, nil, fmt.Errorf("invalid: id is required")
	}
	if body.Name == "" {
		return nil, nil, fmt.Errorf("invalid: name is required")
	}
	u, err := endpointPath("ledgers/%s", id)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(u, http.MethodPut, body)
	if err != nil {
		return nil, nil, err
	}

	ledger := new(Ledger)
	resp, err := s.client.CallWithRetry(req, ledger)
	if err != nil {
		return nil, resp, err
	}

	return ledger, resp, nil
}

func (s *LedgerService) Filter(body FilterParams) (*FilterResponse, *http.Response, error) {
	req, err := s.client.NewRequest("ledgers/filter", http.MethodPost, body)
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

func NewLedgerService(c ClientInterface) *LedgerService {
	return &LedgerService{client: c}
}
