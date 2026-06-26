//go:build integration

// Integration tests for issue #68 — LedgerBalance.Create track_fund_lineage + allocation_strategy.
// Requires Blnk Core running at http://localhost:5001 and BLNK_API_KEY in the environment.
//
// Run:
//
//	export BLNK_API_KEY=your_key
//	go test -tags=integration -v ./integration/... -run Issue68
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue68_CreateBalance_WithLineageFields(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	dob := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)

	ledger, resp, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{Name: "Lineage Create " + suffix})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	identity, resp, err := client.Identity.Create(blnkgo.Identity{
		IdentityType: blnkgo.Individual,
		FirstName:    "Lineage",
		LastName:     "Balance",
		EmailAddress: fmt.Sprintf("issue68-%s@example.com", suffix),
		PhoneNumber:  "1234567890",
		Category:     "customer",
		Street:       "123 Main St",
		Country:      "USA",
		State:        "CA",
		PostCode:     "90001",
		City:         "Los Angeles",
		DOB:          &dob,
		Gender:       "Male",
		Nationality:  "American",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, identity.IdentityId)

	bal, resp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID:           ledger.LedgerID,
		IdentityID:         identity.IdentityId,
		Currency:           "USD",
		TrackFundLineage:   true,
		AllocationStrategy: blnkgo.AllocationStrategyPROPORTIONAL,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.True(t, bal.TrackFundLineage)
	require.Equal(t, blnkgo.AllocationStrategyPROPORTIONAL, bal.AllocationStrategy)
	require.Equal(t, identity.IdentityId, bal.IdentityID)

	got, resp, err := client.LedgerBalance.Get(bal.BalanceID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.True(t, got.TrackFundLineage)
	require.Equal(t, blnkgo.AllocationStrategyPROPORTIONAL, got.AllocationStrategy)
}

func TestIssue68_CreateBalance_AllocationStrategyOnly(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	ledger, resp, err := client.Ledger.Create(blnkgo.CreateLedgerRequest{Name: "Alloc Only " + suffix})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	bal, resp, err := client.LedgerBalance.Create(blnkgo.CreateLedgerBalanceRequest{
		LedgerID:           ledger.LedgerID,
		Currency:           "USD",
		AllocationStrategy: blnkgo.AllocationStrategyFIFO,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, blnkgo.AllocationStrategyFIFO, bal.AllocationStrategy)
}
