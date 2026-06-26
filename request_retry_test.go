package blnkgo

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNormalizeRetryCount(t *testing.T) {
	require.Equal(t, 1, normalizeRetryCount(0))
	require.Equal(t, 1, normalizeRetryCount(-1))
	require.Equal(t, 3, normalizeRetryCount(3))
}

func TestNormalizeRetryDelay(t *testing.T) {
	require.Equal(t, defaultRetryDelay, normalizeRetryDelay(0))
	require.Equal(t, 500*time.Millisecond, normalizeRetryDelay(500*time.Millisecond))
}

func TestRetryDelayForAttempt(t *testing.T) {
	require.Equal(t, 2*time.Second, retryDelayForAttempt(1, 2*time.Second))
	require.Equal(t, 4*time.Second, retryDelayForAttempt(2, 2*time.Second))
}

func TestIsRetryableHTTPMethod(t *testing.T) {
	require.True(t, isRetryableHTTPMethod(http.MethodGet))
	require.False(t, isRetryableHTTPMethod(http.MethodPost))
	require.False(t, isRetryableHTTPMethod(http.MethodPut))
}

func TestIsRetryableHTTPStatus(t *testing.T) {
	require.False(t, isRetryableHTTPStatus(http.StatusBadRequest))
	require.True(t, isRetryableHTTPStatus(http.StatusInternalServerError))
	require.True(t, isRetryableHTTPStatus(http.StatusBadGateway))
}
