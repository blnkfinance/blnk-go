//go:build integration

// Integration tests for the Core 0.15.9 alignment (PR #151).
// Requires Blnk Core 0.15.9+ and BLNK_API_KEY.
//
// Run: go test -tags=integration -v ./integration/... -run Core0159
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func uniqueSuffix() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }

// newMonitorFixture creates a GL balance and a level-triggered monitor on it
// with balance > 100.00 and the description "original".
func newMonitorFixture(t *testing.T, client *blnkgo.Client) (balanceID, monitorID string) {
	t.Helper()
	bal, _, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID:  blnkgo.GeneralLedgerID,
		Currency:  "USD",
		Indicator: "@SdkGoMon" + uniqueSuffix(),
	})
	require.NoError(t, err)

	mon, resp, err := client.BalanceMonitor.Create(blnkgo.MonitorData{
		BalanceID:   bal.BalanceID,
		Description: "original",
		Trigger:     blnkgo.MonitorTriggerLevel,
		Condition: blnkgo.MonitorCondition{
			Field:     "balance",
			Operator:  blnkgo.OperatorGreaterThan,
			Value:     100,
			Precision: 100,
		},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, mon.MonitorID)
	return bal.BalanceID, mon.MonitorID
}

func getMonitor(t *testing.T, client *blnkgo.Client, id string) *blnkgo.MonitorDataResp {
	t.Helper()
	m, _, err := client.BalanceMonitor.Get(id)
	require.NoError(t, err)
	return m
}

// requireOriginalCondition checks the fixture's threshold survived an update.
func requireOriginalCondition(t *testing.T, m *blnkgo.MonitorDataResp) {
	t.Helper()
	assert.Equal(t, "balance", m.Condition.Field)
	assert.Equal(t, blnkgo.OperatorGreaterThan, m.Condition.Operator)
	assert.EqualValues(t, 100, m.Condition.Value)
	assert.EqualValues(t, 100, m.Condition.Precision)
	require.NotNil(t, m.Condition.PreciseValue)
	assert.Equal(t, "10000", m.Condition.PreciseValue.String())
}

func TestCore0159_MonitorUpdatePartial_KeepsOmittedFields(t *testing.T) {
	client := newIntegrationClient(t)
	balanceID, monitorID := newMonitorFixture(t, client)

	created := getMonitor(t, client, monitorID)
	requireOriginalCondition(t, created)
	assert.Equal(t, blnkgo.MonitorTriggerLevel, created.Trigger)

	t.Run("description only keeps the condition, trigger and balance", func(t *testing.T) {
		desc := "renamed"
		out, resp, err := client.BalanceMonitor.UpdatePartial(monitorID, blnkgo.UpdateMonitorRequest{Description: &desc})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		assert.False(t, out.PrecisionUnchanged)

		m := getMonitor(t, client, monitorID)
		assert.Equal(t, "renamed", m.Description)
		assert.Equal(t, balanceID, m.BalanceID)
		assert.Equal(t, blnkgo.MonitorTriggerLevel, m.Trigger)
		requireOriginalCondition(t, m)
	})

	t.Run("trigger only keeps the condition", func(t *testing.T) {
		_, _, err := client.BalanceMonitor.UpdatePartial(monitorID, blnkgo.UpdateMonitorRequest{Trigger: blnkgo.MonitorTriggerEdge})
		require.NoError(t, err)

		m := getMonitor(t, client, monitorID)
		assert.Equal(t, blnkgo.MonitorTriggerEdge, m.Trigger)
		assert.Equal(t, "renamed", m.Description)
		requireOriginalCondition(t, m)
	})

	t.Run("precision without value is ignored and reported", func(t *testing.T) {
		p := int64(1000)
		out, _, err := client.BalanceMonitor.UpdatePartial(monitorID, blnkgo.UpdateMonitorRequest{
			Condition: &blnkgo.UpdateMonitorCondition{Precision: &p},
		})
		require.NoError(t, err)
		assert.True(t, out.PrecisionUnchanged)
		requireOriginalCondition(t, getMonitor(t, client, monitorID))
	})

	t.Run("value only moves the threshold and keeps field and operator", func(t *testing.T) {
		v := 250.0
		_, _, err := client.BalanceMonitor.UpdatePartial(monitorID, blnkgo.UpdateMonitorRequest{
			Condition: &blnkgo.UpdateMonitorCondition{Value: &v},
		})
		require.NoError(t, err)

		m := getMonitor(t, client, monitorID)
		assert.Equal(t, "balance", m.Condition.Field)
		assert.Equal(t, blnkgo.OperatorGreaterThan, m.Condition.Operator)
		assert.EqualValues(t, 250, m.Condition.Value)
		require.NotNil(t, m.Condition.PreciseValue)
		assert.Equal(t, "25000", m.Condition.PreciseValue.String())
	})

	t.Run("explicit zero value sets the threshold to zero", func(t *testing.T) {
		zero := 0.0
		_, _, err := client.BalanceMonitor.UpdatePartial(monitorID, blnkgo.UpdateMonitorRequest{
			Condition: &blnkgo.UpdateMonitorCondition{Value: &zero},
		})
		require.NoError(t, err)

		m := getMonitor(t, client, monitorID)
		assert.Equal(t, "balance", m.Condition.Field)
		assert.Equal(t, blnkgo.OperatorGreaterThan, m.Condition.Operator)
		assert.EqualValues(t, 0, m.Condition.Value)
		assert.EqualValues(t, 100, m.Condition.Precision)
		require.NotNil(t, m.Condition.PreciseValue)
		assert.Equal(t, "0", m.Condition.PreciseValue.String())
		assert.Equal(t, "renamed", m.Description)
	})

	t.Run("empty description clears it", func(t *testing.T) {
		empty := ""
		_, _, err := client.BalanceMonitor.UpdatePartial(monitorID, blnkgo.UpdateMonitorRequest{Description: &empty})
		require.NoError(t, err)
		assert.Equal(t, "", getMonitor(t, client, monitorID).Description)
	})
}

// The deprecated Update sends a full MonitorData. A description-only call
// marshals an empty condition with value 0; Core 0.15.9 must keep the stored
// threshold rather than reset it.
func TestCore0159_MonitorLegacyUpdate_DoesNotResetThreshold(t *testing.T) {
	client := newIntegrationClient(t)
	_, monitorID := newMonitorFixture(t, client)

	_, resp, err := client.BalanceMonitor.Update(monitorID, blnkgo.MonitorData{Description: "legacy rename"})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	m := getMonitor(t, client, monitorID)
	assert.Equal(t, "legacy rename", m.Description)
	assert.Equal(t, blnkgo.MonitorTriggerLevel, m.Trigger)
	requireOriginalCondition(t, m)
}

func TestCore0159_InflightReferenceIsIdempotent(t *testing.T) {
	client := newIntegrationClient(t)
	src := "@SdkGoInfSrc" + uniqueSuffix()
	dst := "@SdkGoInfDst" + uniqueSuffix()

	hold, _, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount: 50, Precision: 100, Currency: "USD",
			Reference: "inf-" + uniqueSuffix(), Description: "SDK inflight reference",
			Source: src, Destination: dst, SkipQueue: true,
		},
		AllowOverdraft: true,
		Inflight:       true,
	})
	require.NoError(t, err)
	require.Equal(t, blnkgo.PryTransactionStatusInFlight, hold.Status)

	ref := "settle-" + uniqueSuffix()
	first, resp, err := client.Transaction.Update(hold.TransactionID, blnkgo.UpdateStatus{
		Status: blnkgo.InflightStatusCommit, Amount: 10, Reference: ref, SkipQueue: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, blnkgo.PryTransactionStatusApplied, first.Status)

	again, resp, err := client.Transaction.Update(hold.TransactionID, blnkgo.UpdateStatus{
		Status: blnkgo.InflightStatusCommit, Amount: 10, Reference: ref, SkipQueue: true,
	})
	require.NoError(t, err, "repeating the same reference, action and amount replays")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, first.TransactionID, again.TransactionID, "the replay returns the original child")

	_, resp, err = client.Transaction.Update(hold.TransactionID, blnkgo.UpdateStatus{
		Status: blnkgo.InflightStatusCommit, Amount: 20, Reference: ref, SkipQueue: true,
	})
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	apiErr, ok := blnkgo.AsApiErrorResponse(err)
	require.True(t, ok)
	require.NotNil(t, apiErr.ErrorDetail)
	assert.Equal(t, blnkgo.ErrorCodeTxnDuplicateReference, apiErr.ErrorDetail.Code)
}

func TestCore0159_OverdraftLimitIsMajorUnits(t *testing.T) {
	client := newIntegrationClient(t)
	src := "@SdkGoOdSrc" + uniqueSuffix()
	dst := "@SdkGoOdDst" + uniqueSuffix()

	post := func(amount float64) (*blnkgo.Transaction, error) {
		txn, _, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
			ParentTransaction: blnkgo.ParentTransaction{
				Amount: amount, Precision: 100, Currency: "USD",
				Reference: "od-" + uniqueSuffix(), Description: "SDK overdraft limit",
				Source: src, Destination: dst, SkipQueue: true,
			},
			AllowOverdraft: true,
			OverdraftLimit: 5, // $5.00
		})
		return txn, err
	}

	within, err := post(3)
	require.NoError(t, err)
	assert.Equal(t, blnkgo.PryTransactionStatusApplied, within.Status, "-$3.00 is within a $5.00 limit")
	assert.EqualValues(t, 5, within.OverdraftLimit, "the response echoes overdraft_limit")

	beyond, err := post(3)
	if err == nil {
		assert.Equal(t, blnkgo.PryTransactionStatusRejected, beyond.Status, "-$6.00 exceeds a $5.00 limit")
	}

	bal, _, err := client.LedgerBalance.GetByIndicator(src, "USD")
	require.NoError(t, err)
	assert.Equal(t, "-300", bal.Balance.String(), "only the first post moved the balance")
}

func TestCore0159_TransactionResponseFields(t *testing.T) {
	client := newIntegrationClient(t)
	txn, _, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount: 1.5, Precision: 100, Currency: "USD",
			Reference: "fields-" + uniqueSuffix(), Description: "SDK response fields",
			Source: "@SdkGoFldSrc" + uniqueSuffix(), Destination: "@SdkGoFldDst" + uniqueSuffix(),
			SkipQueue: true,
		},
		AllowOverdraft: true,
	})
	require.NoError(t, err)

	assert.NotEmpty(t, txn.Hash, "create returns the hash-chain digest")
	assert.Equal(t, "1.5", txn.AmountString)
	assert.True(t, txn.AllowOverdraft, "create echoes allow_overdraft")
	assert.False(t, txn.Inflight)
	require.NotNil(t, txn.ScheduledFor)
	assert.True(t, txn.ScheduledFor.IsZero(), "an unset date arrives as the zero time")

	got, _, err := client.Transaction.Get(txn.TransactionID)
	require.NoError(t, err)
	assert.Equal(t, txn.Hash, got.Hash, "Get returns the same hash")
	// Core stores allow_overdraft in meta_data, not a column, so a read
	// returns the field as false. This pins that behaviour.
	assert.False(t, got.AllowOverdraft)
	assert.Equal(t, true, got.MetaData["allow_overdraft"])
}

func TestCore0159_BalanceIncludeAndPagination(t *testing.T) {
	client := newIntegrationClient(t)
	s := uniqueSuffix()

	ledger, _, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{Name: "sdk-go-include-" + s})
	require.NoError(t, err)
	identity, _, err := client.Identity.Create(blnkgo.Identity{
		IdentityType: blnkgo.Individual, FirstName: "Include", LastName: "Test",
		EmailAddress: "include-" + s + "@example.com", PhoneNumber: "1234567890",
		Category: "customer", Street: "1 Main St", Country: "USA", State: "CA", PostCode: "90001", City: "LA",
	})
	require.NoError(t, err)
	bal, _, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID: ledger.LedgerID, IdentityID: identity.IdentityId, Currency: "USD",
	})
	require.NoError(t, err)

	plain, _, err := client.LedgerBalance.Get(bal.BalanceID)
	require.NoError(t, err)
	assert.Nil(t, plain.Identity)
	assert.Nil(t, plain.Ledger)

	full, _, err := client.LedgerBalance.Get(bal.BalanceID, &blnkgo.GetBalanceRequest{
		Include: []blnkgo.BalanceInclude{blnkgo.BalanceIncludeIdentity, blnkgo.BalanceIncludeLedger},
	})
	require.NoError(t, err)
	require.NotNil(t, full.Ledger)
	assert.Equal(t, ledger.Name, full.Ledger.Name)
	require.NotNil(t, full.Identity)
	assert.Equal(t, "Include", full.Identity.FirstName)

	for name, list := range map[string]func(*blnkgo.ListOptions) (int, error){
		"ledgers": func(o *blnkgo.ListOptions) (int, error) {
			r, _, err := client.Ledger.ListWithOptions(o)
			return len(r), err
		},
		"balances": func(o *blnkgo.ListOptions) (int, error) {
			r, _, err := client.LedgerBalance.ListWithOptions(o)
			return len(r), err
		},
		"transactions": func(o *blnkgo.ListOptions) (int, error) {
			r, _, err := client.Transaction.ListWithOptions(o)
			return len(r), err
		},
	} {
		n, err := list(&blnkgo.ListOptions{Limit: 1})
		require.NoError(t, err, name)
		assert.LessOrEqual(t, n, 1, "%s honours limit=1", name)
	}
}

func TestCore0159_UpdateMetadataReturnsMergedMetadata(t *testing.T) {
	client := newIntegrationClient(t)
	ledger, _, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{
		Name:     "sdk-go-meta-" + uniqueSuffix(),
		MetaData: blnkgo.MetaData{"keep": "yes"},
	})
	require.NoError(t, err)

	out, _, err := client.Metadata.UpdateMetadata(ledger.LedgerID, blnkgo.UpdateMetaDataRequest{
		MetaData: blnkgo.MetaData{"tier": "gold"},
	})
	require.NoError(t, err)
	assert.Equal(t, "gold", out.MetaData["tier"])
	assert.Equal(t, "yes", out.MetaData["keep"], "Core merges onto existing metadata")
}
