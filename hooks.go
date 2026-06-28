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

func NewHooksService(client ClientInterface) *HooksService {
	return &HooksService{client: client}
}
