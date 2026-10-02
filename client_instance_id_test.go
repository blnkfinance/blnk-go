package blnkgo_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestCloudProxyBaseURL(t *testing.T) {
	u, err := url.Parse(blnkgo.CloudProxyBaseURL)
	require.NoError(t, err)
	require.Equal(t, "https", u.Scheme)
	require.Equal(t, "api.cloud.blnkfinance.com", u.Host)
	require.True(t, strings.HasPrefix(u.Path, "/proxy"))
}

func TestNewRequest_OmitsInstanceIDWhenUnset(t *testing.T) {
	u, err := url.Parse("http://localhost:5001/")
	require.NoError(t, err)
	client := blnkgo.NewClient(u, nil)

	req, err := client.NewRequest("ledgers", http.MethodPost, blnkgo.CreateLedgerRequest{Name: "Core"})
	require.NoError(t, err)
	require.Empty(t, req.URL.Query().Get("instance_id"))
}

func TestNewRequest_AddsInstanceIDQueryParam(t *testing.T) {
	u, err := url.Parse(blnkgo.CloudProxyBaseURL)
	require.NoError(t, err)
	client := blnkgo.NewClient(u, nil, blnkgo.WithInstanceID("instance_abc"))

	req, err := client.NewRequest("ledgers", http.MethodPost, blnkgo.CreateLedgerRequest{Name: "Proxy Ledger"})
	require.NoError(t, err)
	require.Equal(t, "instance_abc", req.URL.Query().Get("instance_id"))
	require.Equal(t, "/proxy/ledgers", req.URL.Path)
	require.Equal(t, http.MethodPost, req.Method)
}

func TestNewRequest_MergesInstanceIDWithExistingQuery(t *testing.T) {
	u, err := url.Parse("https://api.cloud.blnkfinance.com/proxy/")
	require.NoError(t, err)
	client := blnkgo.NewClient(u, nil, blnkgo.WithInstanceID("instance_hist"))

	endpoint := "balances/bln_1/at?timestamp=2025-08-30T01:38:30Z&from_source=true"
	req, err := client.NewRequest(endpoint, http.MethodGet, nil)
	require.NoError(t, err)
	q := req.URL.Query()
	require.Equal(t, "instance_hist", q.Get("instance_id"))
	require.Equal(t, "2025-08-30T01:38:30Z", q.Get("timestamp"))
	require.Equal(t, "true", q.Get("from_source"))
}

func TestNewRequest_MergesInstanceIDWithGETQueryOptions(t *testing.T) {
	u, err := url.Parse(blnkgo.CloudProxyBaseURL)
	require.NoError(t, err)
	client := blnkgo.NewClient(u, nil, blnkgo.WithInstanceID("instance_hooks"))

	req, err := client.NewRequest("hooks", http.MethodGet, blnkgo.ListHooksOptions{Type: blnkgo.HookTypePreTransaction})
	require.NoError(t, err)
	q := req.URL.Query()
	require.Equal(t, "instance_hooks", q.Get("instance_id"))
	require.Equal(t, string(blnkgo.HookTypePreTransaction), q.Get("type"))
}

func TestSetInstanceID_AppliesToSubsequentRequests(t *testing.T) {
	u, err := url.Parse("http://localhost:5001/")
	require.NoError(t, err)
	client := blnkgo.NewClient(u, nil)

	req, err := client.NewRequest("ledgers", http.MethodGet, nil)
	require.NoError(t, err)
	require.Empty(t, req.URL.Query().Get("instance_id"))

	client.SetInstanceID("instance_set")
	req, err = client.NewRequest("ledgers", http.MethodGet, nil)
	require.NoError(t, err)
	require.Equal(t, "instance_set", req.URL.Query().Get("instance_id"))
}

func TestNewFileUploadRequest_AddsInstanceIDQueryParam(t *testing.T) {
	u, err := url.Parse(blnkgo.CloudProxyBaseURL)
	require.NoError(t, err)
	client := blnkgo.NewClient(u, nil, blnkgo.WithInstanceID("instance_upload"))

	tmp := t.TempDir()
	path := filepath.Join(tmp, "upload.csv")
	require.NoError(t, os.WriteFile(path, []byte("id,amount\n1,10\n"), 0o644))

	req, err := client.NewFileUploadRequest("reconciliation/upload", "file", path, "upload.csv", map[string]string{"source": "bank"})
	require.NoError(t, err)
	require.Equal(t, "instance_upload", req.URL.Query().Get("instance_id"))
	require.Equal(t, "/proxy/reconciliation/upload", req.URL.Path)
}

func TestLedgerCreate_SendsInstanceIDOnProxyRequest(t *testing.T) {
	var gotURL *url.URL
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"ledger_id": "ldg_1",
			"name":      "My Integration Ledger",
		})
	}))
	defer server.Close()

	base, err := url.Parse(server.URL + "/proxy/")
	require.NoError(t, err)
	apiKey := "cloud_key"
	client := blnkgo.NewClient(base, &apiKey, blnkgo.WithInstanceID("instance_live"))

	ledger, resp, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{Name: "My Integration Ledger"})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "ldg_1", ledger.LedgerID)
	require.Equal(t, "instance_live", gotURL.Query().Get("instance_id"))
	require.Equal(t, "/proxy/ledgers", gotURL.Path)
}

func TestLedgerGet_SendsInstanceIDWithEmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "instance_get", r.URL.Query().Get("instance_id"))
		require.Equal(t, http.MethodGet, r.Method)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Empty(t, body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"ledger_id": "ldg_get",
			"name":      "Existing",
		})
	}))
	defer server.Close()

	base, err := url.Parse(server.URL + "/")
	require.NoError(t, err)
	client := blnkgo.NewClient(base, nil, blnkgo.WithInstanceID("instance_get"), blnkgo.WithTimeout(2*time.Second))

	ledger, _, err := client.Ledger.Get("ldg_get")
	require.NoError(t, err)
	require.Equal(t, "ldg_get", ledger.LedgerID)
}

func TestNewRequest_LeadingSlashEndpointKeepsBasePath(t *testing.T) {
	u, err := url.Parse(blnkgo.CloudProxyBaseURL)
	require.NoError(t, err)
	client := blnkgo.NewClient(u, nil, blnkgo.WithInstanceID("instance_slash"))

	req, err := client.NewRequest("/ledgers", http.MethodPost, blnkgo.CreateLedgerRequest{Name: "Slash"})
	require.NoError(t, err)
	require.Equal(t, "instance_slash", req.URL.Query().Get("instance_id"))
	require.Equal(t, "/proxy/ledgers", req.URL.Path)
}

func TestNewFileUploadRequest_LeadingSlashEndpointKeepsBasePath(t *testing.T) {
	u, err := url.Parse(blnkgo.CloudProxyBaseURL)
	require.NoError(t, err)
	client := blnkgo.NewClient(u, nil, blnkgo.WithInstanceID("instance_slash_upload"))
	path := writeTempUpload(t)

	req, err := client.NewFileUploadRequest("/reconciliation/upload", "file", path, "upload.csv", map[string]string{"source": "bank"})
	require.NoError(t, err)
	require.Equal(t, "instance_slash_upload", req.URL.Query().Get("instance_id"))
	require.Equal(t, "/proxy/reconciliation/upload", req.URL.Path)
}

func TestNewFileUploadRequest_SetBaseURLWithAndWithoutTrailingSlash(t *testing.T) {
	initial, err := url.Parse("https://api.cloud.blnkfinance.com/proxy/")
	require.NoError(t, err)
	client := blnkgo.NewClient(initial, nil, blnkgo.WithInstanceID("instance_upload_base"))
	path := writeTempUpload(t)

	withoutSlash, err := url.Parse("https://api.cloud.blnkfinance.com/proxy")
	require.NoError(t, err)
	client.SetBaseURL(withoutSlash)
	req, err := client.NewFileUploadRequest("reconciliation/upload", "file", path, "upload.csv", map[string]string{"source": "bank"})
	require.NoError(t, err)
	require.Equal(t, "https", req.URL.Scheme)
	require.Equal(t, "api.cloud.blnkfinance.com", req.URL.Host)
	require.Equal(t, "/proxy/reconciliation/upload", req.URL.Path)
	require.Equal(t, "instance_upload_base", req.URL.Query().Get("instance_id"))

	withSlash, err := url.Parse("https://api.cloud.blnkfinance.com/proxy/")
	require.NoError(t, err)
	client.SetBaseURL(withSlash)
	req, err = client.NewFileUploadRequest("reconciliation/upload", "file", path, "upload.csv", map[string]string{"source": "bank"})
	require.NoError(t, err)
	require.Equal(t, "/proxy/reconciliation/upload", req.URL.Path)
	require.Equal(t, "instance_upload_base", req.URL.Query().Get("instance_id"))
}

func TestSetInstanceID_ConcurrentWithRequests(t *testing.T) {
	u, err := url.Parse("http://localhost:5001/")
	require.NoError(t, err)
	client := blnkgo.NewClient(u, nil, blnkgo.WithInstanceID("instance_a"))

	ids := []string{"instance_a", "instance_b", "instance_c"}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			client.SetInstanceID(ids[i%len(ids)])
		}(i)
		go func() {
			defer wg.Done()
			req, err := client.NewRequest("ledgers", http.MethodGet, nil)
			require.NoError(t, err)
			got := req.URL.Query().Get("instance_id")
			require.Contains(t, ids, got)
		}()
	}
	wg.Wait()
}

func writeTempUpload(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	path := filepath.Join(tmp, "upload.csv")
	require.NoError(t, os.WriteFile(path, []byte("id,amount\n1,10\n"), 0o644))
	return path
}
