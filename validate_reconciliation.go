package blnkgo

import "fmt"

// ValidateRunInstantReconData mirrors Core HTTP checks for POST /reconciliation/start-instant.
// Per-field external transaction rules are enforced by Core, not the SDK.
func ValidateRunInstantReconData(data RunInstantReconData) error {
	if len(data.ExternalTransactions) == 0 {
		return fmt.Errorf("external_transactions must be a non-empty array")
	}
	if len(data.ExternalTransactions) > MaxInstantReconciliationItems {
		return fmt.Errorf("too many external_transactions; max is %d", MaxInstantReconciliationItems)
	}
	if data.Strategy == "" {
		return fmt.Errorf("strategy is required")
	}
	if len(data.MatchingRuleIDs) == 0 {
		return fmt.Errorf("matching_rule_ids must be a non-empty array")
	}
	return nil
}
