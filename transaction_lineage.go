package blnkgo

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
)

// LineageFundAllocation is the per-provider fund allocation for a debit transaction
// from a lineage-enabled balance.
type LineageFundAllocation struct {
	Provider string   `json:"provider"`
	Amount   *big.Int `json:"amount"`
}

// TransactionLineage is the fund lineage view for a transaction, including
// provider-level fund allocation and internal shadow transactions.
type TransactionLineage struct {
	TransactionID      string                  `json:"transaction_id"`
	FundAllocation     []LineageFundAllocation `json:"fund_allocation,omitempty"`
	ShadowTransactions []Transaction           `json:"shadow_transactions"`
}

func (a *LineageFundAllocation) UnmarshalJSON(data []byte) error {
	var raw struct {
		Provider string          `json:"provider"`
		Amount   json.RawMessage `json:"amount"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	amount, err := unmarshalBigIntJSON(raw.Amount)
	if err != nil {
		return err
	}

	a.Provider = raw.Provider
	a.Amount = amount
	return nil
}

func (s *TransactionService) GetLineage(transactionID string) (*TransactionLineage, *http.Response, error) {
	if transactionID == "" {
		return nil, nil, fmt.Errorf("transactionID is required")
	}

	u := fmt.Sprintf("transactions/%s/lineage", transactionID)
	req, err := s.client.NewRequest(u, http.MethodGet, nil)
	if err != nil {
		return nil, nil, err
	}

	lineage := new(TransactionLineage)
	resp, err := s.client.CallWithRetry(req, lineage)
	if err != nil {
		return nil, resp, err
	}

	return lineage, resp, nil
}
