package blnkgo

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func redirectVia(t *testing.T, from string) []*http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, from, nil)
	require.NoError(t, err)
	return []*http.Request{req}
}

func redirectTarget(t *testing.T, to string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, to, nil)
	require.NoError(t, err)
	return req
}

func TestRefuseCrossHostRedirect_Policy(t *testing.T) {
	cases := []struct {
		name    string
		from    string
		to      string
		wantErr string
	}{
		{"same scheme and host", "https://api.cloud.blnkfinance.com/proxy/ledgers", "https://api.cloud.blnkfinance.com/proxy/ledgers/", ""},
		{"same host case-insensitive", "https://api.cloud.blnkfinance.com/a", "https://API.CLOUD.BLNKFINANCE.COM/b", ""},
		{"https to http downgrade", "https://api.cloud.blnkfinance.com/proxy/ledgers", "http://api.cloud.blnkfinance.com/proxy/ledgers", "scheme changed"},
		{"http to https upgrade", "http://core.internal:5001/ledgers", "https://core.internal:5001/ledgers", "scheme changed"},
		{"other host", "https://api.cloud.blnkfinance.com/proxy/ledgers", "https://evil.example/proxy/ledgers", "cross-host redirect"},
		{"other port", "https://core.internal:5001/ledgers", "https://core.internal:5002/ledgers", "cross-host redirect"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := refuseCrossHostRedirect(redirectTarget(t, tc.to), redirectVia(t, tc.from))
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestRefuseCrossHostRedirect_StopsAfterTenHops(t *testing.T) {
	via := make([]*http.Request, 0, 10)
	for i := 0; i < 10; i++ {
		via = append(via, redirectVia(t, "https://api.cloud.blnkfinance.com/hop")[0])
	}
	err := refuseCrossHostRedirect(redirectTarget(t, "https://api.cloud.blnkfinance.com/hop"), via)
	require.Error(t, err)
	require.Contains(t, err.Error(), "10 redirects")
}

// Drives the real http.Client redirect machinery: the first hop answers with a
// 302 to the same host over plain http. The downgrade target must never be
// requested, so X-Blnk-Key is never written to a plaintext connection.
func TestCallWithRetry_RefusesHTTPSToHTTPDowngrade(t *testing.T) {
	var plaintextSawKey string
	var hops []string
	u := mustParseURL(t, "https://api.cloud.blnkfinance.com/proxy/")
	key := "cloud_key"
	client := NewClient(u, &key, WithInstanceID("instance_tls"))
	client.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		hops = append(hops, req.URL.String())
		if req.URL.Scheme == "http" {
			plaintextSawKey = req.Header.Get("X-Blnk-Key")
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{}`)), Header: make(http.Header)}, nil
		}
		h := make(http.Header)
		h.Set("Location", "http://api.cloud.blnkfinance.com/proxy/ledgers?instance_id=instance_tls")
		return &http.Response{StatusCode: http.StatusFound, Body: io.NopCloser(strings.NewReader("")), Header: h}, nil
	})

	req, err := client.NewRequest("ledgers", http.MethodGet, nil)
	require.NoError(t, err)
	resp, err := client.CallWithRetry(req, &map[string]any{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "scheme changed")
	require.Nil(t, resp)
	require.Empty(t, plaintextSawKey)
	require.Len(t, hops, 1)
	require.Equal(t, "https", mustParseURL(t, hops[0]).Scheme)
}
