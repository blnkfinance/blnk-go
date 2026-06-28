package blnkgo

import (
	"fmt"
	"strings"
)

func isValidHookType(t HookType) bool {
	switch t {
	case HookTypePreTransaction, HookTypePostTransaction:
		return true
	default:
		return false
	}
}

// ValidateCreateHookRequest performs client-side checks before POST /hooks.
func ValidateCreateHookRequest(body CreateHookRequest) error {
	if strings.TrimSpace(body.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(body.URL) == "" {
		return fmt.Errorf("url is required")
	}
	if !isValidHookType(body.Type) {
		return fmt.Errorf("type must be PRE_TRANSACTION or POST_TRANSACTION")
	}
	if body.Timeout <= 0 {
		return fmt.Errorf("timeout must be a positive number")
	}
	if body.RetryCount < 0 {
		return fmt.Errorf("retry_count must be a non-negative number")
	}
	return nil
}

// ValidateUpdateHookRequest performs client-side checks before PUT /hooks/{id}.
func ValidateUpdateHookRequest(hookID string, body UpdateHookRequest) error {
	if err := ValidateHookID(hookID); err != nil {
		return err
	}
	return ValidateCreateHookRequest(body)
}

// ValidateHookID performs client-side checks before hook operations that require an ID.
func ValidateHookID(hookID string) error {
	if strings.TrimSpace(hookID) == "" {
		return fmt.Errorf("hook id is required")
	}
	return nil
}

// ValidateListHooksOptions performs client-side checks before GET /hooks.
func ValidateListHooksOptions(options *ListHooksOptions) error {
	if options == nil {
		return nil
	}
	if options.Type != "" && !isValidHookType(options.Type) {
		return fmt.Errorf("type must be PRE_TRANSACTION or POST_TRANSACTION")
	}
	return nil
}
