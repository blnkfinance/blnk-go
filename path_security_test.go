package blnkgo_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingServer returns a client pointed at a local server that records
// the raw request URI of everything it receives.
func recordingServer(t *testing.T, basePath string, opts ...blnkgo.ClientOption) (*blnkgo.Client, func() []string) {
	t.Helper()
	var mu sync.Mutex
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen = append(seen, r.Method+" "+r.RequestURI)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL + basePath)
	require.NoError(t, err)
	apiKey := "test_key"
	opts = append([]blnkgo.ClientOption{blnkgo.WithRetry(0)}, opts...)
	client := blnkgo.NewClient(u, &apiKey, opts...)
	return client, func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), seen...)
	}
}

// Every SDK method that puts a caller-supplied value into the request path.
func idMethods(c *blnkgo.Client) map[string]func(id string) error {
	return map[string]func(id string) error{
		"Ledger.Get": func(id string) error { _, _, err := c.Ledger.Get(id); return err },
		"Ledger.Update": func(id string) error {
			_, _, err := c.Ledger.Update(id, blnkgo.UpdateLedgerRequest{Name: "x"})
			return err
		},
		"LedgerBalance.Get": func(id string) error {
			_, _, err := c.LedgerBalance.Get(id)
			return err
		},
		"LedgerBalance.GetByIndicator(indicator)": func(id string) error {
			_, _, err := c.LedgerBalance.GetByIndicator(id, "USD")
			return err
		},
		"LedgerBalance.GetByIndicator(currency)": func(id string) error {
			_, _, err := c.LedgerBalance.GetByIndicator("@World", id)
			return err
		},
		"LedgerBalance.GetHistorical": func(id string) error {
			_, _, err := c.LedgerBalance.GetHistorical(id, time.Now(), false)
			return err
		},
		"LedgerBalance.GetLineage": func(id string) error { _, _, err := c.LedgerBalance.GetLineage(id); return err },
		"LedgerBalance.UpdateIdentity": func(id string) error {
			_, _, err := c.LedgerBalance.UpdateIdentity(id, blnkgo.UpdateBalanceIdentityRequest{IdentityID: "idt_1"})
			return err
		},
		"Transaction.Get":        func(id string) error { _, _, err := c.Transaction.Get(id); return err },
		"Transaction.GetLineage": func(id string) error { _, _, err := c.Transaction.GetLineage(id); return err },
		"Transaction.Refund":     func(id string) error { _, _, err := c.Transaction.Refund(id); return err },
		"Transaction.RefundDryRun": func(id string) error {
			_, _, err := c.Transaction.RefundDryRun(id)
			return err
		},
		"Transaction.Update": func(id string) error {
			_, _, err := c.Transaction.Update(id, blnkgo.UpdateStatus{Status: blnkgo.InflightStatusCommit})
			return err
		},
		"Transaction.UpdateDryRun": func(id string) error {
			_, _, err := c.Transaction.UpdateDryRun(id, blnkgo.UpdateStatus{Status: blnkgo.InflightStatusCommit})
			return err
		},
		"Identity.Get":    func(id string) error { _, _, err := c.Identity.Get(id); return err },
		"Identity.Update": func(id string) error { _, _, err := c.Identity.Update(id, &blnkgo.Identity{}); return err },
		"Identity.Delete": func(id string) error { _, _, err := c.Identity.Delete(id); return err },
		"Identity.TokenizeField(id)": func(id string) error {
			_, _, err := c.Identity.TokenizeField(id, "first_name")
			return err
		},
		"Identity.TokenizeField(field)": func(id string) error {
			_, _, err := c.Identity.TokenizeField("idt_1", id)
			return err
		},
		"Identity.Tokenize": func(id string) error {
			_, _, err := c.Identity.Tokenize(id, blnkgo.TokenizeRequest{Fields: []blnkgo.TokenizableIdentityField{"first_name"}})
			return err
		},
		"Identity.GetTokenizedFields": func(id string) error { _, _, err := c.Identity.GetTokenizedFields(id); return err },
		"Identity.DetokenizeField(id)": func(id string) error {
			_, _, err := c.Identity.DetokenizeField(id, "first_name")
			return err
		},
		"Identity.DetokenizeField(field)": func(id string) error {
			_, _, err := c.Identity.DetokenizeField("idt_1", id)
			return err
		},
		"Identity.Detokenize": func(id string) error {
			_, _, err := c.Identity.Detokenize(id, blnkgo.DetokenizeRequest{Fields: []blnkgo.TokenizableIdentityField{"first_name"}})
			return err
		},
		"Hooks.Get": func(id string) error { _, _, err := c.Hooks.Get(id); return err },
		"Hooks.Update": func(id string) error {
			_, _, err := c.Hooks.Update(id, blnkgo.UpdateHookRequest{Name: "x"})
			return err
		},
		"Hooks.Delete": func(id string) error { _, _, err := c.Hooks.Delete(id); return err },
		"ApiKeys.Delete": func(id string) error {
			_, err := c.ApiKeys.Delete(id, &blnkgo.DeleteApiKeysOptions{Owner: "owner"})
			return err
		},
		"BalanceMonitor.Get":    func(id string) error { _, _, err := c.BalanceMonitor.Get(id); return err },
		"BalanceMonitor.Update": func(id string) error { _, _, err := c.BalanceMonitor.Update(id, blnkgo.MonitorData{}); return err },
		"BalanceMonitor.Delete": func(id string) error { _, _, err := c.BalanceMonitor.Delete(id); return err },
		"BalanceMonitor.ListByBalanceID": func(id string) error {
			_, _, err := c.BalanceMonitor.ListByBalanceID(id)
			return err
		},
		"Reconciliation.Get": func(id string) error { _, _, err := c.Reconciliation.Get(id); return err },
		"Reconciliation.UpdateMatchingRule": func(id string) error {
			_, _, err := c.Reconciliation.UpdateMatchingRule(id, blnkgo.Matcher{})
			return err
		},
		"Reconciliation.DeleteMatchingRule": func(id string) error {
			_, _, err := c.Reconciliation.DeleteMatchingRule(id)
			return err
		},
		"Metadata.UpdateMetadata": func(id string) error {
			_, _, err := c.Metadata.UpdateMetadata(id, blnkgo.UpdateMetaDataRequest{})
			return err
		},
		"Search.SearchDocument": func(id string) error {
			_, _, err := c.Search.SearchDocument(blnkgo.SearchParams{Q: "*"}, blnkgo.ResourceType(id))
			return err
		},
	}
}

var hostileSegments = []string{
	"",
	".",
	"..",
	"../backup",
	"../search/reindex",
	"../api-keys/api_key_victim",
	"%2e%2e/%2e%2e/secret",
	"%2E%2E",
	"..%2Fbackup",
	"ldg_1/../../backup",
	"ldg_1?limit=1",
	"ldg_1#frag",
	`..\backup`,
	"ldg_1\nX-Injected: 1",
}

// No hostile ID may produce a request: the SDK must refuse before sending the
// API key anywhere.
func TestPathSegments_HostileIDsNeverReachTheServer(t *testing.T) {
	for _, basePath := range []string{"/", "/proxy/"} {
		var opts []blnkgo.ClientOption
		if basePath == "/proxy/" {
			opts = append(opts, blnkgo.WithInstanceID("instance_test"))
		}
		client, seen := recordingServer(t, basePath, opts...)
		for name, call := range idMethods(client) {
			for _, id := range hostileSegments {
				err := call(id)
				assert.Errorf(t, err, "%s base=%s id=%q should be rejected", name, basePath, id)
			}
		}
		assert.Empty(t, seen(), "base=%s: hostile IDs reached the server", basePath)
	}
}

// The scanner's proof-of-concept calls, checked by name.
func TestPathSegments_ScannerPoCs(t *testing.T) {
	client, seen := recordingServer(t, "/")

	_, _, err := client.Transaction.Refund("../search/reindex")
	assert.Error(t, err)
	_, _, err = client.Identity.Get("../backup")
	assert.Error(t, err)
	_, _, err = client.Hooks.Delete("../api-keys/api_key_victim")
	assert.Error(t, err)
	assert.Empty(t, seen())

	proxyClient, proxySeen := recordingServer(t, "/proxy/", blnkgo.WithInstanceID("instance_test"))
	_, _, err = proxyClient.Ledger.Get("%2e%2e/%2e%2e/secret")
	assert.Error(t, err)
	assert.Empty(t, proxySeen())
}

// Ordinary IDs are unchanged on the wire, and unusual-but-safe characters
// are escaped rather than rejected.
func TestPathSegments_ValidIDsAreSentEscaped(t *testing.T) {
	client, seen := recordingServer(t, "/")

	_, _, err := client.Ledger.Get("ldg_0b9e8a7c-1234")
	require.NoError(t, err)
	_, _, err = client.LedgerBalance.GetByIndicator("@World", "USD")
	require.NoError(t, err)
	_, _, err = client.Identity.Get("idt with space")
	require.NoError(t, err)
	_, _, err = client.Metadata.UpdateMetadata("txn_1", blnkgo.UpdateMetaDataRequest{})
	require.NoError(t, err)
	_, _, err = client.Search.SearchDocument(blnkgo.SearchParams{Q: "*"}, blnkgo.Transactions)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"GET /ledgers/ldg_0b9e8a7c-1234",
		"GET /balances/indicator/@World/currency/USD",
		"GET /identities/idt%20with%20space",
		"POST /txn_1/metadata",
		"POST /search/transactions",
	}, seen())
}

// References are free-form: "/" and "%" are escaped, dot segments rejected.
func TestGetByReference_EscapesButRejectsDotSegments(t *testing.T) {
	client, seen := recordingServer(t, "/")

	_, _, err := client.Transaction.GetByReference("order/2025/%41")
	require.NoError(t, err)
	_, _, err = client.Transaction.GetByReference("..")
	assert.Error(t, err)
	_, _, err = client.Transaction.GetByReference(".")
	assert.Error(t, err)

	assert.Equal(t, []string{"GET /transactions/reference/order%2F2025%2F%2541"}, seen())
}

// Percent-encoded dot segments passed straight to NewRequest are refused,
// not just plain "../".
func TestNewRequest_RejectsEncodedDotSegments(t *testing.T) {
	u, err := url.Parse(blnkgo.CloudProxyBaseURL)
	require.NoError(t, err)
	apiKey := "cloud_key"
	client := blnkgo.NewClient(u, &apiKey, blnkgo.WithInstanceID("instance_traverse"))

	for _, endpoint := range []string{
		"ledgers/%2e%2e/%2e%2e/data/lake",
		"ledgers/%2E%2E/secret",
		"ledgers/.%2e/secret",
		"ledgers/./x",
	} {
		req, err := client.NewRequest(endpoint, http.MethodGet, nil)
		assert.Errorf(t, err, "endpoint %q", endpoint)
		assert.Nil(t, req)
		if err != nil {
			assert.Contains(t, err.Error(), "escapes")
		}
	}
}

// Timestamps with a positive offset keep their "+" instead of arriving as a
// space.
func TestGetHistorical_EscapesTimestamp(t *testing.T) {
	client, seen := recordingServer(t, "/")
	ts := time.Date(2025, 8, 30, 1, 38, 30, 0, time.FixedZone("+02:00", 2*60*60))

	_, _, err := client.LedgerBalance.GetHistorical("bln_1", ts, true)
	require.NoError(t, err)

	got := seen()
	require.Len(t, got, 1)
	u, err := url.Parse(got[0][len("GET "):])
	require.NoError(t, err)
	assert.Equal(t, "/balances/bln_1/at", u.Path)
	assert.Equal(t, "2025-08-30T01:38:30+02:00", u.Query().Get("timestamp"))
	assert.Equal(t, "true", u.Query().Get("from_source"))
}
