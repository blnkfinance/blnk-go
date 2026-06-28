package blnkgo

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type ApiKeysService service

// CreateApiKeyRequest is the body for POST /api-keys.
type CreateApiKeyRequest struct {
	Name      string    `json:"name"`
	Owner     string    `json:"owner"`
	Scopes    []string  `json:"scopes"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ApiKeyResponse is returned when an API key is created, listed, or fetched.
type ApiKeyResponse struct {
	ApiKeyID   string   `json:"api_key_id"`
	Key        string   `json:"key"`
	Name       string   `json:"name"`
	OwnerID    string   `json:"owner_id"`
	Scopes     []string `json:"scopes"`
	ExpiresAt  string   `json:"expires_at"`
	CreatedAt  string   `json:"created_at"`
	LastUsedAt string   `json:"last_used_at"`
	IsRevoked  bool     `json:"is_revoked"`
	RevokedAt  string   `json:"revoked_at,omitempty"`
}

// Create issues a new scoped API key.
func (s *ApiKeysService) Create(body CreateApiKeyRequest) (*ApiKeyResponse, *http.Response, error) {
	if err := ValidateCreateApiKeyRequest(body); err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("api-keys", http.MethodPost, body)
	if err != nil {
		return nil, nil, err
	}

	apiKeyResp := new(ApiKeyResponse)
	resp, err := s.client.CallWithRetry(req, apiKeyResp)
	if err != nil {
		return nil, resp, err
	}

	return apiKeyResp, resp, nil
}

// ListApiKeysOptions filters GET /api-keys results.
type ListApiKeysOptions struct {
	Owner string `url:"owner,omitempty"`
}

// List returns API keys for an owner. Pass ListApiKeysOptions with Owner when using a master key.
func (s *ApiKeysService) List(options *ListApiKeysOptions) ([]ApiKeyResponse, *http.Response, error) {
	if err := ValidateListApiKeysOptions(options); err != nil {
		return nil, nil, err
	}

	var query interface{}
	if options != nil {
		query = options
	}

	req, err := s.client.NewRequest("api-keys", http.MethodGet, query)
	if err != nil {
		return nil, nil, err
	}

	var keys []ApiKeyResponse
	resp, err := s.client.CallWithRetry(req, &keys)
	if err != nil {
		return nil, resp, err
	}

	return keys, resp, nil
}

// DeleteApiKeysOptions sets optional query params for DELETE /api-keys/{id}.
type DeleteApiKeysOptions struct {
	Owner string
}

// Delete revokes an API key by ID. Pass DeleteApiKeysOptions with Owner when using a master key.
func (s *ApiKeysService) Delete(apiKeyID string, options *DeleteApiKeysOptions) (*http.Response, error) {
	if err := ValidateDeleteApiKeys(apiKeyID, options); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("api-keys/%s", apiKeyID)
	if options != nil && options.Owner != "" {
		endpoint += "?owner=" + url.QueryEscape(options.Owner)
	}

	req, err := s.client.NewRequest(endpoint, http.MethodDelete, nil)
	if err != nil {
		return nil, err
	}

	var discard struct{}
	resp, err := s.client.CallWithRetry(req, &discard)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func NewApiKeysService(client ClientInterface) *ApiKeysService {
	return &ApiKeysService{client: client}
}
