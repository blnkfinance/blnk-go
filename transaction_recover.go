package blnkgo

import (
	"fmt"
	"net/http"
	"net/url"
)

// RecoverQueueRequest holds optional query parameters for POST /transactions/recover.
// Threshold is a Go duration string (e.g. "5m", "1h"). Zero value omits the query
// param and uses the Core default (2 minutes).
type RecoverQueueRequest struct {
	Threshold string `json:"-"`
}

// RecoverQueueResponse is returned after manually recovering stuck queued transactions.
type RecoverQueueResponse struct {
	Recovered int    `json:"recovered"`
	Threshold string `json:"threshold"`
}

func (s *TransactionService) RecoverQueue(body RecoverQueueRequest) (*RecoverQueueResponse, *http.Response, error) {
	if err := ValidateRecoverQueue(body); err != nil {
		return nil, nil, err
	}

	endpoint := "transactions/recover"
	if body.Threshold != "" {
		endpoint = fmt.Sprintf("transactions/recover?threshold=%s", url.QueryEscape(body.Threshold))
	}

	req, err := s.client.NewRequest(endpoint, http.MethodPost, nil)
	if err != nil {
		return nil, nil, err
	}

	response := new(RecoverQueueResponse)
	resp, err := s.client.CallWithRetry(req, response)
	if err != nil {
		return nil, resp, err
	}

	return response, resp, nil
}
