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

func TestValidateUpdateHookRequest(t *testing.T) {
	valid := CreateHookRequest{
		Name:       "Pre-transaction validation",
		URL:        "https://api.example.com/validate",
		Type:       HookTypePreTransaction,
		Active:     true,
		Timeout:    30,
		RetryCount: 3,
	}

	assert.NoError(t, ValidateUpdateHookRequest("hk_test_123", valid))
	assert.Error(t, ValidateUpdateHookRequest("", valid))
	assert.Error(t, ValidateUpdateHookRequest("   ", valid))
	assert.Error(t, ValidateUpdateHookRequest("hk_test_123", CreateHookRequest{Name: valid.Name, Type: valid.Type, Active: true, Timeout: 30, RetryCount: 3}))
}

func TestValidateHookID(t *testing.T) {
	assert.NoError(t, ValidateHookID("hk_test_123"))
	assert.Error(t, ValidateHookID(""))
	assert.Error(t, ValidateHookID("   "))
}

func TestValidateListHooksOptions(t *testing.T) {
	assert.NoError(t, ValidateListHooksOptions(nil))
	assert.NoError(t, ValidateListHooksOptions(&ListHooksOptions{}))
	assert.NoError(t, ValidateListHooksOptions(&ListHooksOptions{Type: HookTypePreTransaction}))
	assert.Error(t, ValidateListHooksOptions(&ListHooksOptions{Type: "INVALID"}))
}
