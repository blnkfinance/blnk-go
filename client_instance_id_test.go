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

func TestInstanceID_IsFixedPerClient(t *testing.T) {
	u, err := url.Parse(blnkgo.CloudProxyBaseURL)
	require.NoError(t, err)
	clientA := blnkgo.NewClient(u, nil, blnkgo.WithInstanceID("instance_a"))
	clientB := blnkgo.NewClient(u, nil, blnkgo.WithInstanceID("instance_b"))

	reqA, err := clientA.NewRequest("ledgers", http.MethodGet, nil)
	require.NoError(t, err)
	reqB, err := clientB.NewRequest("ledgers", http.MethodGet, nil)
	require.NoError(t, err)
	require.Equal(t, "instance_a", reqA.URL.Query().Get("instance_id"))
	require.Equal(t, "instance_b", reqB.URL.Query().Get("instance_id"))
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

func TestNewRequest_ConcurrentReadsSameInstanceID(t *testing.T) {
	u, err := url.Parse(blnkgo.CloudProxyBaseURL)
	require.NoError(t, err)
	client := blnkgo.NewClient(u, nil, blnkgo.WithInstanceID("instance_fixed"))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, err := client.NewRequest("ledgers", http.MethodGet, nil)
			require.NoError(t, err)
			require.Equal(t, "instance_fixed", req.URL.Query().Get("instance_id"))
			require.Equal(t, "api.cloud.blnkfinance.com", req.URL.Host)
			require.Equal(t, "/proxy/ledgers", req.URL.Path)
		}()
	}
	wg.Wait()
}

func TestNewRequest_RejectsAbsoluteEndpoint(t *testing.T) {
	u, err := url.Parse(blnkgo.CloudProxyBaseURL)
	require.NoError(t, err)
	apiKey := "cloud_key"
	client := blnkgo.NewClient(u, &apiKey, blnkgo.WithInstanceID("instance_abs"))

	req, err := client.NewRequest("https://example.com/path", http.MethodPost, blnkgo.CreateLedgerRequest{Name: "Abs"})
	require.Error(t, err)
	require.Nil(t, req)
	require.Contains(t, err.Error(), "relative path")
}

func TestNewRequest_RejectsSchemeRelativeEndpoint(t *testing.T) {
	u, err := url.Parse(blnkgo.CloudProxyBaseURL)
	require.NoError(t, err)
	apiKey := "cloud_key"
	client := blnkgo.NewClient(u, &apiKey, blnkgo.WithInstanceID("instance_scheme"))

	req, err := client.NewRequest("//example.com/path", http.MethodGet, nil)
	require.Error(t, err)
	require.Nil(t, req)
	require.Contains(t, err.Error(), "relative path")
}

func TestNewRequest_RejectsUserinfoEndpoint(t *testing.T) {
	u, err := url.Parse(blnkgo.CloudProxyBaseURL)
	require.NoError(t, err)
	apiKey := "cloud_key"
	client := blnkgo.NewClient(u, &apiKey, blnkgo.WithInstanceID("instance_userinfo"))

	req, err := client.NewRequest("//user:pass@example.com/ledgers", http.MethodGet, nil)
	require.Error(t, err)
	require.Nil(t, req)
	require.Contains(t, err.Error(), "relative path")
}

func TestNewRequest_RejectsPathTraversalEndpoint(t *testing.T) {
	u, err := url.Parse(blnkgo.CloudProxyBaseURL)
	require.NoError(t, err)
	apiKey := "cloud_key"
	client := blnkgo.NewClient(u, &apiKey, blnkgo.WithInstanceID("instance_traverse"))

	req, err := client.NewRequest("../data/lake", http.MethodGet, nil)
	require.Error(t, err)
	require.Nil(t, req)
	require.Contains(t, err.Error(), "escapes")
}

func TestNewRequest_OverwritesInjectedInstanceID(t *testing.T) {
	u, err := url.Parse(blnkgo.CloudProxyBaseURL)
	require.NoError(t, err)
	client := blnkgo.NewClient(u, nil, blnkgo.WithInstanceID("instance_client"))

	req, err := client.NewRequest("ledgers?instance_id=instance_attacker&limit=10", http.MethodGet, nil)
	require.NoError(t, err)
	q := req.URL.Query()
	require.Equal(t, "instance_client", q.Get("instance_id"))
	require.Equal(t, []string{"instance_client"}, q["instance_id"])
	require.Equal(t, "10", q.Get("limit"))
}

func TestNewRequest_CloudProxyRequiresInstanceID(t *testing.T) {
	u, err := url.Parse(blnkgo.CloudProxyBaseURL)
	require.NoError(t, err)
	client := blnkgo.NewClient(u, nil)

	req, err := client.NewRequest("ledgers", http.MethodGet, nil)
	require.Error(t, err)
	require.Nil(t, req)
	require.Contains(t, err.Error(), "instance_id is required")
}

func TestNewRequest_CloudProxyRejectsNonInstancePrefix(t *testing.T) {
	u, err := url.Parse(blnkgo.CloudProxyBaseURL)
	require.NoError(t, err)
	client := blnkgo.NewClient(u, nil, blnkgo.WithInstanceID("deploy_not_an_instance"))

	req, err := client.NewRequest("ledgers", http.MethodGet, nil)
	require.Error(t, err)
	require.Nil(t, req)
	require.Contains(t, err.Error(), "instance_")
}

func TestNewRequest_AlreadyProxyEndpointDoesNotDouble(t *testing.T) {
	u, err := url.Parse(blnkgo.CloudProxyBaseURL)
	require.NoError(t, err)
	client := blnkgo.NewClient(u, nil, blnkgo.WithInstanceID("instance_proxy"))

	req, err := client.NewRequest("/proxy/ledgers", http.MethodGet, nil)
	require.NoError(t, err)
	require.Equal(t, "/proxy/ledgers", req.URL.Path)
	require.Equal(t, "instance_proxy", req.URL.Query().Get("instance_id"))
}

func TestNewRequest_SetBaseURLWithoutTrailingSlash(t *testing.T) {
	initial, err := url.Parse("http://localhost:5001/")
	require.NoError(t, err)
	client := blnkgo.NewClient(initial, nil)

	withoutSlash, err := url.Parse("https://api.cloud.blnkfinance.com/proxy")
	require.NoError(t, err)
	client.SetBaseURL(withoutSlash)

	req, err := client.NewRequest("ledgers", http.MethodGet, nil)
	require.Error(t, err)
	require.Nil(t, req)

	client = blnkgo.NewClient(initial, nil, blnkgo.WithInstanceID("instance_base"))
	client.SetBaseURL(withoutSlash)
	req, err = client.NewRequest("ledgers", http.MethodGet, nil)
	require.NoError(t, err)
	require.Equal(t, "/proxy/ledgers", req.URL.Path)
	require.Equal(t, "instance_base", req.URL.Query().Get("instance_id"))
}

func TestNewFileUploadRequest_RejectsAbsoluteAndSchemeRelative(t *testing.T) {
	u, err := url.Parse(blnkgo.CloudProxyBaseURL)
	require.NoError(t, err)
	apiKey := "cloud_key"
	client := blnkgo.NewClient(u, &apiKey, blnkgo.WithInstanceID("instance_upload_sec"))
	path := writeTempUpload(t)

	req, err := client.NewFileUploadRequest("https://example.com/upload", "file", path, "upload.csv", map[string]string{"source": "bank"})
	require.Error(t, err)
	require.Nil(t, req)

	req, err = client.NewFileUploadRequest("//example.com/upload", "file", path, "upload.csv", map[string]string{"source": "bank"})
	require.Error(t, err)
	require.Nil(t, req)
}

func TestSetBaseURL_ConcurrentWithRequests(t *testing.T) {
	core, err := url.Parse("http://localhost:5001/")
	require.NoError(t, err)
	proxy, err := url.Parse(blnkgo.CloudProxyBaseURL)
	require.NoError(t, err)
	client := blnkgo.NewClient(core, nil, blnkgo.WithInstanceID("instance_race"))

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				client.SetBaseURL(proxy)
			} else {
				client.SetBaseURL(core)
			}
		}(i)
		go func() {
			defer wg.Done()
			req, err := client.NewRequest("ledgers", http.MethodGet, nil)
			if err != nil {
				require.Contains(t, err.Error(), "instance_id is required")
				return
			}
			require.Contains(t, []string{"localhost:5001", "api.cloud.blnkfinance.com"}, req.URL.Host)
			if req.URL.Host == "api.cloud.blnkfinance.com" {
				require.Equal(t, "instance_race", req.URL.Query().Get("instance_id"))
				require.True(t, strings.HasPrefix(req.URL.Path, "/proxy/"))
			}
		}()
	}
	wg.Wait()
}

func TestNewClient_IgnoresLaterMutationOfCallerURL(t *testing.T) {
	u, err := url.Parse("http://localhost:5001")
	require.NoError(t, err)
	client := blnkgo.NewClient(u, nil)

	u.Host = "evil.example"
	u.Path = "/stolen"

	req, err := client.NewRequest("ledgers", http.MethodGet, nil)
	require.NoError(t, err)
	require.Equal(t, "localhost:5001", req.URL.Host)
	require.Equal(t, "/ledgers", req.URL.Path)
}

func TestClientBaseURLFieldStaysCompatible(t *testing.T) {
	u, err := url.Parse("http://localhost:5001/")
	require.NoError(t, err)
	client := blnkgo.NewClient(u, nil)
	require.NotNil(t, client.BaseURL)
	require.Equal(t, "localhost:5001", client.BaseURL.Host)

	client.BaseURL.Host = "evil.example"
	client.BaseURL.Path = "/stolen"
	req, err := client.NewRequest("ledgers", http.MethodGet, nil)
	require.NoError(t, err)
	require.Equal(t, "localhost:5001", req.URL.Host)
	require.Equal(t, "/ledgers", req.URL.Path)
	require.Equal(t, "localhost:5001", client.BaseURL.Host)

	next, err := url.Parse("http://127.0.0.1:5001/")
	require.NoError(t, err)
	client.BaseURL = next
	req, err = client.NewRequest("balances", http.MethodGet, nil)
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1:5001", req.URL.Host)
	require.Equal(t, "/balances", req.URL.Path)

	next.Host = "evil.example"
	req, err = client.NewRequest("ledgers", http.MethodGet, nil)
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1:5001", req.URL.Host)

	client.BaseURL = nil
	req, err = client.NewRequest("ledgers", http.MethodGet, nil)
	require.Error(t, err)
	require.Nil(t, req)
}

func TestAPIKeyFieldStaysCompatible(t *testing.T) {
	key := "cloud_key"
	u, err := url.Parse("http://localhost:5001/")
	require.NoError(t, err)
	client := blnkgo.NewClient(u, &key)

	key = "replaced-by-caller"
	req, err := client.NewRequest("ledgers", http.MethodGet, nil)
	require.NoError(t, err)
	require.Equal(t, "cloud_key", req.Header.Get("X-Blnk-Key"))

	*client.ApiKey = "mutated-in-place"
	req, err = client.NewRequest("ledgers", http.MethodGet, nil)
	require.NoError(t, err)
	require.Equal(t, "cloud_key", req.Header.Get("X-Blnk-Key"))

	assigned := "assigned_key"
	client.ApiKey = &assigned
	req, err = client.NewRequest("ledgers", http.MethodGet, nil)
	require.NoError(t, err)
	require.Equal(t, "assigned_key", req.Header.Get("X-Blnk-Key"))

	assigned = "changed-after-assign"
	req, err = client.NewRequest("ledgers", http.MethodGet, nil)
	require.NoError(t, err)
	require.Equal(t, "assigned_key", req.Header.Get("X-Blnk-Key"))
}

func TestSetBaseURL_IgnoresLaterMutationOfCallerURL(t *testing.T) {
	initial, err := url.Parse("http://localhost:5001/")
	require.NoError(t, err)
	client := blnkgo.NewClient(initial, nil)

	next, err := url.Parse("http://127.0.0.1:5001")
	require.NoError(t, err)
	client.SetBaseURL(next)
	next.Host = "evil.example"

	req, err := client.NewRequest("ledgers", http.MethodGet, nil)
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1:5001", req.URL.Host)
}

func TestCallerURLMutation_ConcurrentWithRequests(t *testing.T) {
	u, err := url.Parse("http://localhost:5001/")
	require.NoError(t, err)
	client := blnkgo.NewClient(u, nil)

	stop := make(chan struct{})
	mutatorDone := make(chan struct{})
	go func() {
		defer close(mutatorDone)
		for {
			select {
			case <-stop:
				return
			default:
				u.Host = "evil.example"
				u.Path = "/stolen"
				u.Host = "localhost:5001"
				u.Path = "/"
			}
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, err := client.NewRequest("ledgers", http.MethodGet, nil)
			require.NoError(t, err)
			require.Equal(t, "localhost:5001", req.URL.Host)
			require.Equal(t, "/ledgers", req.URL.Path)
		}()
	}
	wg.Wait()
	close(stop)
	<-mutatorDone
}

func writeTempUpload(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	path := filepath.Join(tmp, "upload.csv")
	require.NoError(t, os.WriteFile(path, []byte("id,amount\n1,10\n"), 0o644))
	return path
}
