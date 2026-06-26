package blnkgo_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

type retryLedgerResponse struct {
	LedgerID string `json:"ledger_id"`
	Name     string `json:"name"`
}

func newRetryTestClient(t *testing.T, serverURL string, opts ...blnkgo.ClientOption) *blnkgo.Client {
	t.Helper()
	u, err := url.Parse(serverURL + "/")
	require.NoError(t, err)
	opts = append([]blnkgo.ClientOption{
		blnkgo.WithRetry(3),
		blnkgo.WithRetryDelay(5 * time.Millisecond),
	}, opts...)
	return blnkgo.NewClient(u, nil, opts...)
}

func TestCallWithRetry_RetriesGETOn5xx(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		if attempts.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_ = json.NewEncoder(w).Encode(retryLedgerResponse{LedgerID: "ldg_1", Name: "Retry OK"})
	}))
	defer server.Close()

	client := newRetryTestClient(t, server.URL)
	ledger, resp, err := client.Ledger.Get("ldg_1")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "ldg_1", ledger.LedgerID)
	require.Equal(t, int32(3), attempts.Load())
}

func TestCallWithRetry_DoesNotRetryPOSTOn5xx(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"server error"}`))
	}))
	defer server.Close()

	client := newRetryTestClient(t, server.URL)
	_, resp, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{Name: "No Retry"})
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Equal(t, int32(1), attempts.Load())
}

func TestCallWithRetry_DoesNotRetryGETOn4xx(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	client := newRetryTestClient(t, server.URL)
	_, resp, err := client.Ledger.Get("missing")
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	require.Equal(t, int32(1), attempts.Load())
}

func TestNewRequest_SetsReplayableBody(t *testing.T) {
	u, err := url.Parse("http://localhost:5001/")
	require.NoError(t, err)
	client := blnkgo.NewClient(u, nil)

	req, err := client.NewRequest("ledgers", http.MethodPost, blnkgo.CreateLedgerRequest{Name: "Replay"})
	require.NoError(t, err)
	require.NotNil(t, req.GetBody)

	first, err := req.GetBody()
	require.NoError(t, err)
	defer first.Close()
	var payload1 map[string]string
	require.NoError(t, json.NewDecoder(first).Decode(&payload1))

	second, err := req.GetBody()
	require.NoError(t, err)
	defer second.Close()
	var payload2 map[string]string
	require.NoError(t, json.NewDecoder(second).Decode(&payload2))

	require.Equal(t, payload1, payload2)
	require.Equal(t, "Replay", payload1["name"])
}
