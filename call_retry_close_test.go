package blnkgo

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type trackCloseBody struct {
	io.Reader
	closed *bool
}

func (t *trackCloseBody) Close() error {
	*t.closed = true
	return nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestCallWithRetry_ClosesBodyOnSuccess(t *testing.T) {
	closed := false
	u := mustParseURL(t, "http://example.com/")
	client := NewClient(u, nil)
	client.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: &trackCloseBody{
				Reader: strings.NewReader(`{"ledger_id":"ldg_1","name":"Test"}`),
				closed: &closed,
			},
			Header: make(http.Header),
		}, nil
	})}

	req, err := client.NewRequest("ledgers/ldg_1", http.MethodGet, nil)
	require.NoError(t, err)

	var ledger Ledger
	_, err = client.CallWithRetry(req, &ledger)
	require.NoError(t, err)
	require.True(t, closed)
}

func TestCallWithRetry_ClosesBodyOnTerminalError(t *testing.T) {
	closed := false
	u := mustParseURL(t, "http://example.com/")
	client := NewClient(u, nil)
	client.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body: &trackCloseBody{
				Reader: strings.NewReader(`{"message":"not found"}`),
				closed: &closed,
			},
			Header: make(http.Header),
		}, nil
	})}

	req, err := client.NewRequest("ledgers/missing", http.MethodGet, nil)
	require.NoError(t, err)

	var ledger Ledger
	_, err = client.CallWithRetry(req, &ledger)
	require.Error(t, err)
	require.True(t, closed)
}

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	require.NoError(t, err)
	return u
}
