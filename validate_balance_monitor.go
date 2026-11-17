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

// ValidateBalanceID performs client-side checks before operations that require a balance ID.
func ValidateBalanceID(balanceID string) error {
	if strings.TrimSpace(balanceID) == "" {
		return fmt.Errorf("balance id is required")
	}
	return nil
}

// ValidateMonitorTrigger accepts an empty trigger (Core's default) or one of
// MonitorTriggerEdge and MonitorTriggerLevel.
func ValidateMonitorTrigger(trigger MonitorTrigger) error {
	switch trigger {
	case "", MonitorTriggerEdge, MonitorTriggerLevel:
		return nil
	default:
		return fmt.Errorf("trigger must be %q or %q", MonitorTriggerEdge, MonitorTriggerLevel)
	}
}
