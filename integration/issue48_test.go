//go:build integration

// Integration tests for issue #48 — Identity.DetokenizeField (GET /identities/{id}/detokenize/{field}).
// Requires Blnk Core running at http://localhost:5001 with tokenization enabled.
//
// Run: go test -tags=integration -v ./integration/... -run Issue48
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue48_DetokenizeField(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("issue48-%s@example.com", suffix)

	created, createResp, err := client.Identity.Create(blnkgo.Identity{
		IdentityType: blnkgo.Individual,
		FirstName:    "Detokenize",
		LastName:     "Field",
		EmailAddress: email,
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

	_, tokenizeResp, err := client.Identity.TokenizeField(created.IdentityId, string(blnkgo.TokenizableFieldEmailAddress))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, tokenizeResp.StatusCode)

	detokenized, detokenizeResp, err := client.Identity.DetokenizeField(created.IdentityId, string(blnkgo.TokenizableFieldEmailAddress))
	require.NoError(t, err)
	require.NotNil(t, detokenizeResp)
	require.Equal(t, http.StatusOK, detokenizeResp.StatusCode)
	require.Equal(t, string(blnkgo.TokenizableFieldEmailAddress), detokenized.Field)
	require.Equal(t, email, detokenized.Value)
}

func TestIssue48_DetokenizeField_ValidationError(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Identity.DetokenizeField("", string(blnkgo.TokenizableFieldFirstName))
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "identity id is required")
}

func TestIssue48_DetokenizeField_NotTokenized(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	created, createResp, err := client.Identity.Create(blnkgo.Identity{
		IdentityType: blnkgo.Individual,
		FirstName:    "NotTokenized",
		LastName:     "Field",
		EmailAddress: fmt.Sprintf("issue48-not-%s@example.com", suffix),
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

	_, resp, err := client.Identity.DetokenizeField(created.IdentityId, string(blnkgo.TokenizableFieldFirstName))
	require.Error(t, err)
	require.NotNil(t, resp)
	require.NotEqual(t, http.StatusOK, resp.StatusCode)
}
