//go:build integration

// Integration tests for issue #117 — Identity.Delete (DELETE /identities/{id}).
// Requires Blnk Core 0.15.0+ at http://localhost:5001.
//
// Run: go test -tags=integration -v ./integration/... -run Issue117
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue117_DeleteIdentity(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	created, createResp, err := client.Identity.Create(blnkgo.Identity{
		IdentityType: blnkgo.Individual,
		FirstName:    "Delete",
		LastName:     "Me",
		EmailAddress: fmt.Sprintf("issue117-%s@example.com", suffix),
		PhoneNumber:  "1234567890",
		Category:     "customer",
		Street:       "123 Main St",
		Country:      "USA",
		State:        "CA",
		PostCode:     "90001",
		City:         "Los Angeles",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	deleted, deleteResp, err := client.Identity.Delete(created.IdentityId)
	require.NoError(t, err)
	require.NotNil(t, deleteResp)
	require.Equal(t, http.StatusOK, deleteResp.StatusCode)
	require.Equal(t, "Identity deleted successfully", deleted.Message)

	_, getResp, err := client.Identity.Get(created.IdentityId)
	require.Error(t, err)
	require.NotNil(t, getResp)
	require.NotEqual(t, http.StatusOK, getResp.StatusCode)
}

func TestIssue117_DeleteIdentity_ValidationError(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Identity.Delete("")
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "identity id is required")
}

func TestIssue117_DeleteIdentity_NotFound(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Identity.Delete("idt_nonexistent_issue117")
	require.Error(t, err)
	require.NotNil(t, resp)
	require.NotEqual(t, http.StatusOK, resp.StatusCode)
}
