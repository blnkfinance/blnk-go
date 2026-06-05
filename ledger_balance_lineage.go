package blnkgo

import (
	"fmt"
	"math/big"
	"net/http"
)

// LineageProviderBreakdown is the per-provider fund breakdown for a balance with
// fund lineage tracking enabled.
type LineageProviderBreakdown struct {
	Provider        string   `json:"provider"`
	Amount          *big.Int `json:"amount"`
	Available       *big.Int `json:"available"`
	Spent           *big.Int `json:"spent"`
	ShadowBalanceID string   `json:"shadow_balance_id"`
}

// BalanceLineage is the fund lineage view for a balance, including provider-level
// received, spent, and available amounts (minor units).
type BalanceLineage struct {
	BalanceID          string                     `json:"balance_id"`
	AggregateBalanceID string                     `json:"aggregate_balance_id"`
	TotalWithLineage   *big.Int                   `json:"total_with_lineage"`
	Providers          []LineageProviderBreakdown `json:"providers"`
}

func (s *LedgerBalanceService) GetLineage(balanceID string) (*BalanceLineage, *http.Response, error) {
	if balanceID == "" {
		return nil, nil, fmt.Errorf("invalid: balanceID is required")
	}

	u := fmt.Sprintf("balances/%s/lineage", balanceID)
	req, err := s.client.NewRequest(u, http.MethodGet, nil)
	if err != nil {
		return nil, nil, err
	}

	lineage := new(BalanceLineage)
	resp, err := s.client.CallWithRetry(req, lineage)
	if err != nil {
		return nil, resp, err
	}

	return lineage, resp, nil
}
