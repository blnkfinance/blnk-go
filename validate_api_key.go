package blnkgo

import (
	"fmt"
	"strings"
)

// ValidateCreateApiKeyRequest performs client-side checks before POST /api-keys.
func ValidateCreateApiKeyRequest(body CreateApiKeyRequest) error {
	if strings.TrimSpace(body.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(body.Owner) == "" {
		return fmt.Errorf("owner is required")
	}
	if len(body.Scopes) == 0 {
		return fmt.Errorf("at least one scope must be specified")
	}
	for _, scope := range body.Scopes {
		if strings.TrimSpace(scope) == "" {
			return fmt.Errorf("each scope must be a non-empty string")
		}
	}
	if body.ExpiresAt.IsZero() {
		return fmt.Errorf("expires_at is required")
	}
	return nil
}

// ValidateListApiKeysOptions performs client-side checks before GET /api-keys.
func ValidateListApiKeysOptions(options *ListApiKeysOptions) error {
	if options == nil {
		return nil
	}
	if options.Owner != "" && strings.TrimSpace(options.Owner) == "" {
		return fmt.Errorf("owner must be a non-empty string")
	}
	return nil
}

// ValidateDeleteApiKeys performs client-side checks before DELETE /api-keys/{id}.
func ValidateDeleteApiKeys(apiKeyID string, options *DeleteApiKeysOptions) error {
	if strings.TrimSpace(apiKeyID) == "" {
		return fmt.Errorf("api key id is required")
	}
	if options != nil && options.Owner != "" && strings.TrimSpace(options.Owner) == "" {
		return fmt.Errorf("owner must be a non-empty string")
	}
	return nil
}
