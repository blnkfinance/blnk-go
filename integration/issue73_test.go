//go:build integration

// Integration tests for issue #73 — Search.SearchDocument identities ResourceType.
// Requires Blnk Core running at http://localhost:5001 (0.14.x+; search/identities is not 0.15-only).
//
// Run: go test -tags=integration -v ./integration/... -run Issue73
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func newClientIssue73(t *testing.T) *blnkgo.Client {
	return newIntegrationClient(t)
}

func TestIssue73_SearchIdentities(t *testing.T) {
	client := newClientIssue73(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("search-identities-%s@example.com", suffix)
	dob := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)

	identity, resp, err := client.Identity.Create(blnkgo.Identity{
		IdentityType: blnkgo.Individual,
		FirstName:    "Search",
		LastName:     "Test",
		EmailAddress: email,
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

	// Core returns 201 for POST /search/{resource}; accept any 2xx.
	deadline := time.Now().Add(20 * time.Second)
	var searchResp *blnkgo.SearchResponse
	for {
		searchResp, resp, err = client.Search.SearchDocument(blnkgo.SearchParams{
			Q:       email,
			QueryBy: "email_address",
			Page:    1,
			PerPage: 10,
		}, blnkgo.Identities)
		if err == nil && resp.StatusCode < 300 && searchResp.Found > 0 {
			break
		}
		if time.Now().After(deadline) {
			require.NoError(t, err, "search identities timed out waiting for index")
			require.Less(t, resp.StatusCode, 300, "expected 2xx from search/identities")
			require.Greater(t, searchResp.Found, 0, "expected indexed identity in search results")
		}
		time.Sleep(500 * time.Millisecond)
	}

	require.Less(t, resp.StatusCode, 300)
	require.Greater(t, searchResp.Found, 0)
}
