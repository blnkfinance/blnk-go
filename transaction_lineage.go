package blnkgo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// TransactionLineage is the fund lineage view for a transaction, including
// fund allocation metadata and any shadow transactions created for lineage tracking.
type TransactionLineage struct {
	TransactionID      string                   `json:"transaction_id"`
	FundAllocation     []map[string]interface{} `json:"fund_allocation,omitempty"`
	ShadowTransactions []map[string]interface{} `json:"shadow_transactions"`
}

func (t *TransactionLineage) UnmarshalJSON(data []byte) error {
	type alias TransactionLineage
	aux := &struct {
		ShadowTransactions json.RawMessage `json:"shadow_transactions"`
		*alias
	}{
		alias: (*alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if len(aux.ShadowTransactions) == 0 || string(aux.ShadowTransactions) == "null" {
		t.ShadowTransactions = nil
		return nil
	}

	return json.Unmarshal(aux.ShadowTransactions, &t.ShadowTransactions)
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
