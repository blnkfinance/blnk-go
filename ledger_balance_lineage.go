package blnkgo

import (
	"encoding/json"
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

func unmarshalBigIntJSON(data []byte) (*big.Int, error) {
	if string(data) == "null" {
		return nil, nil
	}

	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		v, ok := new(big.Int).SetString(n.String(), 10)
		if ok {
			return v, nil
		}
	}

	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		v, ok := new(big.Int).SetString(s, 10)
		if ok {
			return v, nil
		}
	}

	return nil, fmt.Errorf("cannot parse big.Int from %s", string(data))
}

func (p *LineageProviderBreakdown) UnmarshalJSON(data []byte) error {
	var raw struct {
		Provider        string          `json:"provider"`
		Amount          json.RawMessage `json:"amount"`
		Available       json.RawMessage `json:"available"`
		Spent           json.RawMessage `json:"spent"`
		ShadowBalanceID string          `json:"shadow_balance_id"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	amount, err := unmarshalBigIntJSON(raw.Amount)
	if err != nil {
		return err
	}
	available, err := unmarshalBigIntJSON(raw.Available)
	if err != nil {
		return err
	}
	spent, err := unmarshalBigIntJSON(raw.Spent)
	if err != nil {
		return err
	}

	p.Provider = raw.Provider
	p.Amount = amount
	p.Available = available
	p.Spent = spent
	p.ShadowBalanceID = raw.ShadowBalanceID
	return nil
}

func (b *BalanceLineage) UnmarshalJSON(data []byte) error {
	var raw struct {
		BalanceID          string                     `json:"balance_id"`
		AggregateBalanceID string                     `json:"aggregate_balance_id"`
		TotalWithLineage   json.RawMessage            `json:"total_with_lineage"`
		Providers          []LineageProviderBreakdown `json:"providers"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	total, err := unmarshalBigIntJSON(raw.TotalWithLineage)
	if err != nil {
		return err
	}

	b.BalanceID = raw.BalanceID
	b.AggregateBalanceID = raw.AggregateBalanceID
	b.TotalWithLineage = total
	b.Providers = raw.Providers
	return nil
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
