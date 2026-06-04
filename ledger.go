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
	MetaData  map[string]interface{} `json:"meta_data,omitempty"`
}

type CreateLedgerRequest struct {
	Name     string                 `json:"name"`
	MetaData map[string]interface{} `json:"meta_data,omitempty"`
}

type UpdateLedgerRequest struct {
	Name string `json:"name"`
}

func (s *LedgerService) List() ([]Ledger, *http.Response, error) {
	req, err := s.client.NewRequest("ledgers", http.MethodGet, nil)
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
	u := fmt.Sprintf("ledgers/%s", id)
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
	u := fmt.Sprintf("ledgers/%s", id)
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
