package blnkgo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCreateHookRequest(t *testing.T) {
	valid := CreateHookRequest{
		Name:       "Pre-transaction validation",
		URL:        "https://api.example.com/validate",
		Type:       HookTypePreTransaction,
		Active:     true,
		Timeout:    30,
		RetryCount: 3,
	}

	assert.NoError(t, ValidateCreateHookRequest(valid))
	assert.Error(t, ValidateCreateHookRequest(CreateHookRequest{URL: valid.URL, Type: valid.Type, Active: true, Timeout: 30, RetryCount: 3}))
	assert.Error(t, ValidateCreateHookRequest(CreateHookRequest{Name: valid.Name, Type: valid.Type, Active: true, Timeout: 30, RetryCount: 3}))
	assert.Error(t, ValidateCreateHookRequest(CreateHookRequest{Name: valid.Name, URL: valid.URL, Active: true, Timeout: 30, RetryCount: 3}))
	assert.Error(t, ValidateCreateHookRequest(CreateHookRequest{Name: valid.Name, URL: valid.URL, Type: "INVALID", Active: true, Timeout: 30, RetryCount: 3}))
	assert.Error(t, ValidateCreateHookRequest(CreateHookRequest{Name: valid.Name, URL: valid.URL, Type: valid.Type, Active: true, Timeout: 0, RetryCount: 3}))
	assert.Error(t, ValidateCreateHookRequest(CreateHookRequest{Name: valid.Name, URL: valid.URL, Type: valid.Type, Active: true, Timeout: 30, RetryCount: -1}))
}
