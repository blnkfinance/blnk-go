package blnkgo_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type capturedRequest struct {
	Method string
	URI    string
	Body   string
}

// replyServer returns a client whose server answers every request with reply
// and records what it received, so tests see the real wire format.
func replyServer(t *testing.T, reply string) (*blnkgo.Client, func() capturedRequest) {
	t.Helper()
	var mu sync.Mutex
	var last capturedRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		last = capturedRequest{Method: r.Method, URI: r.RequestURI, Body: string(body)}
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(reply))
	}))
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	apiKey := "test_key"
	client := blnkgo.NewClient(u, &apiKey, blnkgo.WithRetry(0))
	return client, func() capturedRequest {
		mu.Lock()
		defer mu.Unlock()
		return last
	}
}

func TestUpdateMetadataDecodesCoreResponseKey(t *testing.T) {
	client, _ := replyServer(t, `{"meta_data":{"tier":"gold","limit":5}}`)

	got, _, err := client.Metadata.UpdateMetadata("ldg_1", blnkgo.UpdateMetaDataRequest{
		MetaData: blnkgo.MetaData{"tier": "gold"},
	})
	require.NoError(t, err)
	assert.Equal(t, "gold", got.MetaData["tier"])
	assert.EqualValues(t, 5, got.MetaData["limit"])
}

func TestListWithOptionsSendsPagination(t *testing.T) {
	client, last := replyServer(t, `[]`)

	type listFns struct {
		list        func() error
		withOptions func(*blnkgo.ListOptions) error
	}
	calls := map[string]listFns{
		"/ledgers": {
			func() error { _, _, err := client.Ledger.List(); return err },
			func(o *blnkgo.ListOptions) error { _, _, err := client.Ledger.ListWithOptions(o); return err },
		},
		"/balances": {
			func() error { _, _, err := client.LedgerBalance.List(); return err },
			func(o *blnkgo.ListOptions) error { _, _, err := client.LedgerBalance.ListWithOptions(o); return err },
		},
		"/transactions": {
			func() error { _, _, err := client.Transaction.List(); return err },
			func(o *blnkgo.ListOptions) error { _, _, err := client.Transaction.ListWithOptions(o); return err },
		},
		"/identities": {
			func() error { _, _, err := client.Identity.List(); return err },
			func(o *blnkgo.ListOptions) error { _, _, err := client.Identity.ListWithOptions(o); return err },
		},
	}
	for path, fns := range calls {
		t.Run(path, func(t *testing.T) {
			require.NoError(t, fns.list())
			assert.Equal(t, path, last().URI, "List sends no query")

			require.NoError(t, fns.withOptions(nil))
			assert.Equal(t, path, last().URI, "nil options send no query")

			require.NoError(t, fns.withOptions(&blnkgo.ListOptions{Limit: 50, Offset: 100}))
			assert.Equal(t, path+"?limit=50&offset=100", last().URI)

			require.NoError(t, fns.withOptions(&blnkgo.ListOptions{Limit: 5}))
			assert.Equal(t, path+"?limit=5", last().URI)

			assert.Error(t, fns.withOptions(&blnkgo.ListOptions{Limit: -1}))
			assert.Error(t, fns.withOptions(&blnkgo.ListOptions{Offset: -1}))
		})
	}
}

func TestBalanceMonitorTrigger(t *testing.T) {
	client, last := replyServer(t, `{
		"monitor_id":"mon_1","balance_id":"bln_1","trigger":"level","condition_state":true,
		"condition":{"field":"balance","operator":">","value":10,"precision":100,"precise_value":1000}
	}`)

	got, _, err := client.BalanceMonitor.Create(blnkgo.MonitorData{
		BalanceID: "bln_1",
		Trigger:   blnkgo.MonitorTriggerLevel,
		Condition: blnkgo.MonitorCondition{Field: "balance", Operator: blnkgo.OperatorGreaterThan, Value: 10, Precision: 100},
	})
	require.NoError(t, err)

	var sent map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(last().Body), &sent))
	assert.Equal(t, "level", sent["trigger"])

	assert.Equal(t, blnkgo.MonitorTriggerLevel, got.Trigger)
	require.NotNil(t, got.ConditionState)
	assert.True(t, *got.ConditionState)
	require.NotNil(t, got.Condition.PreciseValue)
	assert.Equal(t, "1000", got.Condition.PreciseValue.String())

	_, _, err = client.BalanceMonitor.Create(blnkgo.MonitorData{BalanceID: "bln_1", Trigger: "sometimes"})
	assert.Error(t, err, "unknown trigger is rejected before sending")
}

func TestBalanceMonitorCreateOmitsEmptyTrigger(t *testing.T) {
	client, last := replyServer(t, `{"monitor_id":"mon_1"}`)

	_, _, err := client.BalanceMonitor.Create(blnkgo.MonitorData{BalanceID: "bln_1"})
	require.NoError(t, err)
	assert.NotContains(t, last().Body, "trigger")
}

func TestBalanceMonitorUpdatePartialSendsOnlySetFields(t *testing.T) {
	client, last := replyServer(t, `{"message":"BalanceMonitor updated successfully. precision was left unchanged because value was not sent with it.","precision_unchanged":true}`)

	desc := "renamed"
	got, _, err := client.BalanceMonitor.UpdatePartial("mon_1", blnkgo.UpdateMonitorRequest{Description: &desc})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, last().Method)
	assert.Equal(t, "/balance-monitors/mon_1", last().URI)
	assert.JSONEq(t, `{"description":"renamed"}`, last().Body)
	assert.True(t, got.PrecisionUnchanged)
	assert.Contains(t, got.Message, "updated successfully")

	zero := 0.0
	_, _, err = client.BalanceMonitor.UpdatePartial("mon_1", blnkgo.UpdateMonitorRequest{
		Condition: &blnkgo.UpdateMonitorCondition{Value: &zero},
	})
	require.NoError(t, err)
	assert.JSONEq(t, `{"condition":{"value":0}}`, last().Body, "an explicit zero threshold is sent")

	empty := ""
	_, _, err = client.BalanceMonitor.UpdatePartial("mon_1", blnkgo.UpdateMonitorRequest{
		Description: &empty,
		Trigger:     blnkgo.MonitorTriggerEdge,
	})
	require.NoError(t, err)
	assert.JSONEq(t, `{"description":"","trigger":"edge"}`, last().Body, "an empty description clears it")

	_, _, err = client.BalanceMonitor.UpdatePartial("", blnkgo.UpdateMonitorRequest{})
	assert.Error(t, err)
	_, _, err = client.BalanceMonitor.UpdatePartial("mon_1", blnkgo.UpdateMonitorRequest{Trigger: "bad"})
	assert.Error(t, err)
}

func TestInflightUpdateSendsReference(t *testing.T) {
	client, last := replyServer(t, `{"transaction_id":"txn_2","status":"APPLIED","reference":"settle_1"}`)

	_, _, err := client.Transaction.Update("txn_1", blnkgo.UpdateStatus{
		Status:    blnkgo.InflightStatusCommit,
		Reference: "settle_1",
	})
	require.NoError(t, err)

	var sent map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(last().Body), &sent))
	assert.Equal(t, "settle_1", sent["reference"])

	_, _, err = client.Transaction.Update("txn_1", blnkgo.UpdateStatus{Status: blnkgo.InflightStatusVoid})
	require.NoError(t, err)
	assert.NotContains(t, last().Body, "reference", "reference is omitted when unset")
}

func TestCreateTransactionSendsOverdraftLimitAndDecodesResponseFields(t *testing.T) {
	client, last := replyServer(t, `{
		"transaction_id":"txn_1","reference":"ref_1","status":"QUEUED","currency":"USD",
		"amount":112.3,"amount_string":"112.30","precision":100,"precise_amount":11230,
		"hash":"abc123","inflight":true,"allow_overdraft":true,"overdraft_limit":500,
		"inflight_expiry_date":"2026-12-01T00:00:00Z","scheduled_for":"2026-11-20T09:00:00Z",
		"inflight_commit_date":"0001-01-01T00:00:00Z","queued":true
	}`)

	txn, _, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount: 112.3, Precision: 100, Reference: "ref_1", Currency: "USD",
			Source: "bln_src", Destination: "bln_dst",
		},
		AllowOverdraft: true,
		OverdraftLimit: 500,
	})
	require.NoError(t, err)

	var sent map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(last().Body), &sent))
	assert.EqualValues(t, 500, sent["overdraft_limit"])

	assert.Equal(t, "abc123", txn.Hash)
	assert.Equal(t, "112.30", txn.AmountString)
	assert.True(t, txn.Inflight)
	assert.True(t, txn.AllowOverdraft)
	assert.EqualValues(t, 500, txn.OverdraftLimit)
	assert.True(t, txn.Queued)
	require.NotNil(t, txn.InflightExpiryDate)
	assert.Equal(t, 2026, txn.InflightExpiryDate.Year())
	require.NotNil(t, txn.ScheduledFor)
	assert.Equal(t, 20, txn.ScheduledFor.Day())
}

func TestCreateTransactionOmitsZeroOverdraftLimit(t *testing.T) {
	client, last := replyServer(t, `{"transaction_id":"txn_1"}`)

	_, _, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount: 1, Precision: 100, Reference: "ref_2", Currency: "USD",
			Source: "bln_src", Destination: "bln_dst",
		},
	})
	require.NoError(t, err)
	assert.NotContains(t, last().Body, "overdraft_limit")
}

func TestGetBalanceInclude(t *testing.T) {
	client, last := replyServer(t, `{
		"balance_id":"bln_1","currency":"USD","ledger_id":"ldg_1","identity_id":"idt_1",
		"identity":{"identity_id":"idt_1","first_name":"Ada","email_address":"ada@example.com"},
		"ledger":{"ledger_id":"ldg_1","name":"Customers"}
	}`)

	bal, _, err := client.LedgerBalance.Get("bln_1", &blnkgo.GetBalanceRequest{
		WithQueued: true,
		Include:    []blnkgo.BalanceInclude{blnkgo.BalanceIncludeIdentity, blnkgo.BalanceIncludeLedger},
	})
	require.NoError(t, err)
	assert.Equal(t, "/balances/bln_1?with_queued=true&include=identity&include=ledger", last().URI)

	require.NotNil(t, bal.Identity)
	assert.Equal(t, "Ada", bal.Identity.FirstName)
	require.NotNil(t, bal.Ledger)
	assert.Equal(t, "Customers", bal.Ledger.Name)

	_, _, err = client.LedgerBalance.Get("bln_1", &blnkgo.GetBalanceRequest{
		Include: []blnkgo.BalanceInclude{"transactions"},
	})
	assert.Error(t, err)
}

func TestFilterSendsLogicalOperator(t *testing.T) {
	client, last := replyServer(t, `{"data":[]}`)

	_, _, err := client.Transaction.Filter(blnkgo.FilterParams{
		Filters: []blnkgo.Filter{
			{Field: "status", Operator: blnkgo.OpEqual, Value: "APPLIED"},
			{Field: "status", Operator: blnkgo.OpEqual, Value: "INFLIGHT"},
		},
		LogicalOperator: blnkgo.LogicalOr,
	})
	require.NoError(t, err)

	var sent map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(last().Body), &sent))
	assert.Equal(t, "or", sent["logical_operator"])

	_, _, err = client.Transaction.Filter(blnkgo.FilterParams{Filters: []blnkgo.Filter{{Field: "status", Operator: blnkgo.OpEqual, Value: "APPLIED"}}})
	require.NoError(t, err)
	assert.NotContains(t, last().Body, "logical_operator")
}

// Core 0.15.7 answers a reused commit/void reference with a different amount
// or action with 409 TXN_DUPLICATE_REFERENCE (api/inflight_reference_test.go).
func TestInflightReferenceConflictSurfacesErrorCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"reference settle_1 already used","error_detail":{"code":"TXN_DUPLICATE_REFERENCE","message":"reference settle_1 already used"}}`))
	}))
	t.Cleanup(srv.Close)
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	apiKey := "test_key"
	client := blnkgo.NewClient(u, &apiKey, blnkgo.WithRetry(0))

	_, resp, err := client.Transaction.Update("txn_1", blnkgo.UpdateStatus{
		Status:    blnkgo.InflightStatusCommit,
		Amount:    10,
		Reference: "settle_1",
	})
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	apiErr, ok := blnkgo.AsApiErrorResponse(err)
	require.True(t, ok, "error should be an *ApiErrorResponse")
	require.NotNil(t, apiErr.ErrorDetail)
	assert.Equal(t, blnkgo.ErrorCodeTxnDuplicateReference, apiErr.ErrorDetail.Code)
}
