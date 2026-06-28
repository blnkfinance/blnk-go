package blnkgo

import "net/http"

type HealthService service

// HealthResponse is returned by GET /health when Blnk Core is reachable.
type HealthResponse struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

// Check verifies that Blnk Core is running and reachable.
func (s *HealthService) Check() (*HealthResponse, *http.Response, error) {
	req, err := s.client.NewRequest("health", http.MethodGet, nil)
	if err != nil {
		return nil, nil, err
	}

	healthResp := new(HealthResponse)
	resp, err := s.client.CallWithRetry(req, healthResp)
	if err != nil {
		return nil, resp, err
	}

	return healthResp, resp, nil
}

func NewHealthService(client ClientInterface) *HealthService {
	return &HealthService{client: client}
}
