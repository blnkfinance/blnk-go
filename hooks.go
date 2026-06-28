package blnkgo

import (
	"fmt"
	"net/http"
)

type HooksService service

// HookType identifies when a webhook runs relative to a transaction.
type HookType string

const (
	HookTypePreTransaction  HookType = "PRE_TRANSACTION"
	HookTypePostTransaction HookType = "POST_TRANSACTION"
)

// CreateHookRequest is the body for POST /hooks.
type CreateHookRequest struct {
	Name       string   `json:"name"`
	URL        string   `json:"url"`
	Type       HookType `json:"type"`
	Active     bool     `json:"active"`
	Timeout    int      `json:"timeout"`
	RetryCount int      `json:"retry_count"`
}

// UpdateHookRequest is the body for PUT /hooks/{id}.
type UpdateHookRequest = CreateHookRequest

// HookResponse is returned when a hook is created, listed, fetched, or updated.
type HookResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Type        HookType `json:"type"`
	Active      bool     `json:"active"`
	Timeout     int      `json:"timeout"`
	RetryCount  int      `json:"retry_count"`
	CreatedAt   string   `json:"created_at"`
	LastRun     string   `json:"last_run"`
	LastSuccess bool     `json:"last_success"`
}

// DeleteHookResponse is returned when a hook is deleted.
type DeleteHookResponse struct {
	Message string `json:"message"`
}

// Create registers a new webhook (master key required).
func (s *HooksService) Create(body CreateHookRequest) (*HookResponse, *http.Response, error) {
	if err := ValidateCreateHookRequest(body); err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("hooks", http.MethodPost, body)
	if err != nil {
		return nil, nil, err
	}

	hookResp := new(HookResponse)
	resp, err := s.client.CallWithRetry(req, hookResp)
	if err != nil {
		return nil, resp, err
	}

	return hookResp, resp, nil
}

// Update modifies an existing webhook by ID (master key required).
func (s *HooksService) Update(hookID string, body UpdateHookRequest) (*HookResponse, *http.Response, error) {
	if err := ValidateUpdateHookRequest(hookID, body); err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(fmt.Sprintf("hooks/%s", hookID), http.MethodPut, body)
	if err != nil {
		return nil, nil, err
	}

	hookResp := new(HookResponse)
	resp, err := s.client.CallWithRetry(req, hookResp)
	if err != nil {
		return nil, resp, err
	}

	return hookResp, resp, nil
}

// Get retrieves a webhook by ID (master key required).
func (s *HooksService) Get(hookID string) (*HookResponse, *http.Response, error) {
	if err := ValidateHookID(hookID); err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(fmt.Sprintf("hooks/%s", hookID), http.MethodGet, nil)
	if err != nil {
		return nil, nil, err
	}

	hookResp := new(HookResponse)
	resp, err := s.client.CallWithRetry(req, hookResp)
	if err != nil {
		return nil, resp, err
	}

	return hookResp, resp, nil
}

// ListHooksOptions filters GET /hooks results.
type ListHooksOptions struct {
	Type HookType `url:"type,omitempty"`
}

// List returns webhooks, optionally filtered by type (master key required).
func (s *HooksService) List(options *ListHooksOptions) ([]HookResponse, *http.Response, error) {
	if err := ValidateListHooksOptions(options); err != nil {
		return nil, nil, err
	}

	var query interface{}
	if options != nil && options.Type != "" {
		query = options
	}

	req, err := s.client.NewRequest("hooks", http.MethodGet, query)
	if err != nil {
		return nil, nil, err
	}

	var hooks []HookResponse
	resp, err := s.client.CallWithRetry(req, &hooks)
	if err != nil {
		return nil, resp, err
	}

	return hooks, resp, nil
}

// Delete removes a webhook by ID (master key required).
func (s *HooksService) Delete(hookID string) (*DeleteHookResponse, *http.Response, error) {
	if err := ValidateHookID(hookID); err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(fmt.Sprintf("hooks/%s", hookID), http.MethodDelete, nil)
	if err != nil {
		return nil, nil, err
	}

	deleteResp := new(DeleteHookResponse)
	resp, err := s.client.CallWithRetry(req, deleteResp)
	if err != nil {
		return nil, resp, err
	}

	return deleteResp, resp, nil
}

func NewHooksService(client ClientInterface) *HooksService {
	return &HooksService{client: client}
}
