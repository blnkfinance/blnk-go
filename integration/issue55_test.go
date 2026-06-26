//go:build integration

// Integration tests for issue #55 — Identity.Filter (POST /identities/filter).
//
// Run:
//
//	export BLNK_API_KEY=your_key
//	go test -tags=integration -v ./integration/... -run Issue55
package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func identitiesFromFilter(t *testing.T, result *blnkgo.FilterResponse) []blnkgo.IdentityResponse {
	t.Helper()
	raw, err := json.Marshal(result.Data)
	require.NoError(t, err)
	var identities []blnkgo.IdentityResponse
	require.NoError(t, json.Unmarshal(raw, &identities))
	return identities
}

func TestIssue55_FilterIdentities_ByEmail(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("issue55-filter-%s@example.com", suffix)

	created, resp, err := client.Identity.Create(blnkgo.Identity{
		IdentityType: blnkgo.Individual,
		FirstName:    "Filter",
		LastName:     "Test",
		EmailAddress: email,
		Category:     "customer",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	result, resp, err := client.Identity.Filter(blnkgo.FilterParams{
		Filters: []blnkgo.Filter{
			{Field: "email_address", Operator: blnkgo.OpEqual, Value: email},
		},
		Limit: 10,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	identities := identitiesFromFilter(t, result)
	require.NotEmpty(t, identities)
	require.Equal(t, created.IdentityId, identities[0].IdentityId)
	require.Equal(t, email, identities[0].EmailAddress)
}

func TestIssue55_FilterIdentities_IncludeCount(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("issue55-count-%s@example.com", suffix)

	_, resp, err := client.Identity.Create(blnkgo.Identity{
		IdentityType: blnkgo.Individual,
		FirstName:    "Count",
		EmailAddress: email,
		Category:     "customer",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	result, resp, err := client.Identity.Filter(blnkgo.FilterParams{
		Filters: []blnkgo.Filter{
			{Field: "email_address", Operator: blnkgo.OpEqual, Value: email},
		},
		IncludeCount: true,
		Limit:        10,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, result.TotalCount)
	require.GreaterOrEqual(t, *result.TotalCount, int64(1))
}
