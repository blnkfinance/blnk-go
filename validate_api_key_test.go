package blnkgo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateCreateApiKeyRequest(t *testing.T) {
	valid := CreateApiKeyRequest{
		Name:      "my-key",
		Owner:     "owner_1",
		Scopes:    []string{"ledgers:read"},
		ExpiresAt: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	assert.NoError(t, ValidateCreateApiKeyRequest(valid))
	assert.Error(t, ValidateCreateApiKeyRequest(CreateApiKeyRequest{Owner: "o", Scopes: []string{"s"}, ExpiresAt: valid.ExpiresAt}))
	assert.Error(t, ValidateCreateApiKeyRequest(CreateApiKeyRequest{Name: "n", Scopes: []string{"s"}, ExpiresAt: valid.ExpiresAt}))
	assert.Error(t, ValidateCreateApiKeyRequest(CreateApiKeyRequest{Name: "n", Owner: "o", ExpiresAt: valid.ExpiresAt}))
	assert.Error(t, ValidateCreateApiKeyRequest(CreateApiKeyRequest{Name: "n", Owner: "o", Scopes: []string{""}, ExpiresAt: valid.ExpiresAt}))
	assert.Error(t, ValidateCreateApiKeyRequest(CreateApiKeyRequest{Name: "n", Owner: "o", Scopes: []string{"s"}}))
}
