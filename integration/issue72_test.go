//go:build integration

// Integration tests for issue #72 — optional caller-supplied identity_id and relaxed create validation.
// Requires Blnk Core running at http://localhost:5001 and BLNK_API_KEY in the environment.
//
// Run:
//
//	export BLNK_API_KEY=your_key
//	go test -tags=integration -v ./integration/... -run Issue72
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func callerIdentityID(suffix string) string {
	n := time.Now().UnixNano()
	return fmt.Sprintf("idt_%08x-0000-4000-8000-%012x", uint32(n>>32), n&0xffffffffffff)
}

func TestIssue72_CreateIdentity_CallerSuppliedID(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	callerID := callerIdentityID(suffix)

	created, resp, err := client.Identity.Create(blnkgo.Identity{
		IdentityID:   callerID,
		IdentityType: blnkgo.Individual,
		FirstName:    "CallerID",
		LastName:     "Test",
		EmailAddress: fmt.Sprintf("issue72-caller-%s@example.com", suffix),
		Category:     "customer",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, callerID, created.IdentityId)

	got, resp, err := client.Identity.Get(callerID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, callerID, got.IdentityId)
	require.Equal(t, "CallerID", got.FirstName)
}

func TestIssue72_CreateIdentity_MinimalPayload(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	created, resp, err := client.Identity.Create(blnkgo.Identity{
		IdentityType: blnkgo.Individual,
		FirstName:    "Minimal",
		EmailAddress: fmt.Sprintf("issue72-minimal-%s@example.com", suffix),
		Category:     "customer",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotEmpty(t, created.IdentityId)
	require.Equal(t, "Minimal", created.FirstName)
}
