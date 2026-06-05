package blnkgo

import (
	"fmt"
	"net/http"
)

type UpdateBalanceIdentityRequest struct {
	IdentityID string `json:"identity_id"`
}

type UpdateBalanceIdentityResponse struct {
	Message string `json:"message"`
}

func (s *LedgerBalanceService) UpdateIdentity(balanceID string, body UpdateBalanceIdentityRequest) (*UpdateBalanceIdentityResponse, *http.Response, error) {
	if balanceID == "" {
		return nil, nil, fmt.Errorf("invalid: balanceID is required")
	}
	if body.IdentityID == "" {
		return nil, nil, fmt.Errorf("invalid: identity_id is required")
	}

	u := fmt.Sprintf("balances/%s/identity", balanceID)
	req, err := s.client.NewRequest(u, http.MethodPut, body)
	if err != nil {
		return nil, nil, err
	}

	response := new(UpdateBalanceIdentityResponse)
	resp, err := s.client.CallWithRetry(req, response)
	if err != nil {
		return nil, resp, err
	}

	return response, resp, nil
}
