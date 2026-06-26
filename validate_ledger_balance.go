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
	return nil
}

func normalizeAllocationStrategy(strategy AllocationStrategy) AllocationStrategy {
	if strategy == "" {
		return strategy
	}
	return AllocationStrategy(strings.ToUpper(string(strategy)))
}
