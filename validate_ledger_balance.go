package blnkgo

import (
	"fmt"
	"strings"
)

// AllocationStrategy controls how tagged provider funds are spent when fund lineage is enabled.
type AllocationStrategy string

const (
	AllocationStrategyFIFO         AllocationStrategy = "FIFO"
	AllocationStrategyLIFO         AllocationStrategy = "LIFO"
	AllocationStrategyPROPORTIONAL AllocationStrategy = "PROPORTIONAL"
)

func isValidAllocationStrategy(strategy AllocationStrategy) bool {
	switch strategy {
	case AllocationStrategyFIFO, AllocationStrategyLIFO, AllocationStrategyPROPORTIONAL:
		return true
	default:
		return false
	}
}

func ValidateCreateLedgerBalance(b CreateLedgerBalanceRequest) error {
	if b.TrackFundLineage && b.IdentityID == "" {
		return fmt.Errorf("identity_id is required when track_fund_lineage is enabled")
	}
	if b.AllocationStrategy != "" && !isValidAllocationStrategy(b.AllocationStrategy) {
		return fmt.Errorf("allocation_strategy must be one of FIFO, LIFO, or PROPORTIONAL")
	}
	if b.Indicator != "" {
		if !strings.HasPrefix(b.Indicator, "@") {
			return fmt.Errorf("indicator must start with @")
		}
		if strings.ContainsAny(b.Indicator, " \t\n\r") {
			return fmt.Errorf("indicator must not contain spaces")
		}
		if b.LedgerID != GeneralLedgerID {
			return fmt.Errorf("indicator is only valid when ledger_id is %s", GeneralLedgerID)
		}
	}
	return nil
}

func normalizeAllocationStrategy(strategy AllocationStrategy) AllocationStrategy {
	if strategy == "" {
		return strategy
	}
	return AllocationStrategy(strings.ToUpper(string(strategy)))
}
