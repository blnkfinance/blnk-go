package blnkgo

import (
	"fmt"
	"math"
	"strings"
)

func ValidateRunInstantReconData(data RunInstantReconData) error {
	if len(data.ExternalTransactions) == 0 {
		return fmt.Errorf("external_transactions must be a non-empty array")
	}
	if len(data.ExternalTransactions) > MaxInstantReconciliationItems {
		return fmt.Errorf("too many external_transactions; max is %d", MaxInstantReconciliationItems)
	}
	for _, txn := range data.ExternalTransactions {
		if err := validateExternalTransaction(txn); err != nil {
			return err
		}
	}
	if !isValidReconciliationStrategy(data.Strategy) {
		return fmt.Errorf("strategy must be one of: one_to_one, one_to_many, many_to_one")
	}
	if len(data.MatchingRuleIDs) == 0 {
		return fmt.Errorf("matching_rule_ids must be a non-empty array")
	}
	for _, ruleID := range data.MatchingRuleIDs {
		if strings.TrimSpace(ruleID) == "" {
			return fmt.Errorf("each matching_rule_id must be a valid string")
		}
	}
	return nil
}

func validateExternalTransaction(txn ExternalTransaction) error {
	if strings.TrimSpace(txn.ID) == "" ||
		strings.TrimSpace(txn.Reference) == "" ||
		strings.TrimSpace(txn.Currency) == "" ||
		strings.TrimSpace(txn.Description) == "" ||
		strings.TrimSpace(txn.Source) == "" {
		return fmt.Errorf("each external transaction must include id, amount, reference, currency, description, date, and source")
	}
	if txn.Date.IsZero() {
		return fmt.Errorf("each external transaction must include id, amount, reference, currency, description, date, and source")
	}
	if math.IsNaN(txn.Amount) || math.IsInf(txn.Amount, 0) {
		return fmt.Errorf("each external transaction must include id, amount, reference, currency, description, date, and source")
	}
	return nil
}

func isValidReconciliationStrategy(strategy ReconciliationStrategy) bool {
	switch strategy {
	case ReconciliationStrategyOneToOne, ReconciliationStrategyOneToMany, ReconciliationStrategyManyToOne:
		return true
	default:
		return false
	}
}
