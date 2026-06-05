package blnkgo

import (
	"fmt"
	"net/http"
)

type CreateBalanceSnapshotRequest struct {
	// BatchSize controls how many balances are processed per batch.
	// Zero omits the query parameter and uses the server default (1000).
	BatchSize int `json:"-"`
}

type CreateBalanceSnapshotResponse struct {
	Message string `json:"message"`
}

func (s *LedgerBalanceService) CreateSnapshot(body CreateBalanceSnapshotRequest) (*CreateBalanceSnapshotResponse, *http.Response, error) {
	if body.BatchSize < 0 {
		return nil, nil, fmt.Errorf("invalid: batch_size must be positive")
	}

	endpoint := "balances-snapshots"
	if body.BatchSize > 0 {
		endpoint = fmt.Sprintf("balances-snapshots?batch_size=%d", body.BatchSize)
	}

	req, err := s.client.NewRequest(endpoint, http.MethodPost, nil)
	if err != nil {
		return nil, nil, err
	}

	response := new(CreateBalanceSnapshotResponse)
	resp, err := s.client.CallWithRetry(req, response)
	if err != nil {
		return nil, resp, err
	}

	return response, resp, nil
}
