package blnkgo

import (
	"net/http"
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

func NewApiKeysService(client ClientInterface) *ApiKeysService {
	return &ApiKeysService{client: client}
}
