package blnkgo

import (
	"errors"
	"net/http"
	"time"
)

const defaultRetryDelay = 2 * time.Second

func normalizeRetryCount(count int) int {
	if count < 1 {
		return 1
	}
	return count
}

func normalizeRetryDelay(delay time.Duration) time.Duration {
	if delay <= 0 {
		return defaultRetryDelay
	}
	return delay
}

func isRetryableHTTPMethod(method string) bool {
	return method == http.MethodGet
}

func isRetryableHTTPStatus(statusCode int) bool {
	return statusCode >= http.StatusInternalServerError
}

func retryDelayForAttempt(attempt int, baseDelay time.Duration) time.Duration {
	if attempt < 1 {
		return baseDelay
	}
	return baseDelay * time.Duration(attempt)
}

func resetRequestBody(req *http.Request) error {
	if req.GetBody == nil {
		return nil
	}
	body, err := req.GetBody()
	if err != nil {
		return err
	}
	req.Body = body
	return nil
}

func isRetryableNetworkError(err error) bool {
	if err == nil {
		return false
	}
	// Timeouts are not retried to avoid duplicate mutating calls when the server may have processed the request.
	if errors.Is(err, http.ErrHandlerTimeout) {
		return false
	}
	var netErr interface{ Timeout() bool }
	if errors.As(err, &netErr) && netErr.Timeout() {
		return false
	}
	return true
}
