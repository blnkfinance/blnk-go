package blnkgo

import (
	"fmt"
	"strings"
)

// ValidateMonitorID performs client-side checks before balance monitor operations that require an ID.
func ValidateMonitorID(monitorID string) error {
	if strings.TrimSpace(monitorID) == "" {
		return fmt.Errorf("monitor id is required")
	}
	return nil
}
