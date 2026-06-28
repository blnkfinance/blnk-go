//go:build integration

// Integration tests for issue #46 — Identity.Tokenize (POST /identities/{id}/tokenize).
// Requires Blnk Core running at http://localhost:5001 with tokenization enabled.
//
// Run: go test -tags=integration -v ./integration/... -run Issue46
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue46_Tokenize(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	created, createResp, err := client.Identity.Create(blnkgo.Identity{
		IdentityType: blnkgo.Individual,
		FirstName:    "Multi",
		LastName:     "Tokenize",
		EmailAddress: fmt.Sprintf("issue46-%s@example.com", suffix),
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
	require.NotEmpty(t, created.IdentityId)

	tokenized, tokenizeResp, err := client.Identity.Tokenize(created.IdentityId, blnkgo.TokenizeRequest{
		Fields: []blnkgo.TokenizableIdentityField{
			blnkgo.TokenizableFieldFirstName,
			blnkgo.TokenizableFieldLastName,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, tokenizeResp)
	require.Equal(t, http.StatusOK, tokenizeResp.StatusCode)
	require.Equal(t, "Fields tokenized successfully", tokenized.Message)
}

func TestIssue46_Tokenize_ValidationError(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Identity.Tokenize("", blnkgo.TokenizeRequest{
		Fields: []blnkgo.TokenizableIdentityField{blnkgo.TokenizableFieldFirstName},
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "identity id is required")
}

func TestIssue46_Tokenize_EmptyFields(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Identity.Tokenize("idt_test_123", blnkgo.TokenizeRequest{})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "at least one field must be specified")
}
