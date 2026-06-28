//go:build integration

// Integration tests for issue #49 — Identity.Detokenize (POST /identities/{id}/detokenize).
// Requires Blnk Core running at http://localhost:5001 with tokenization enabled.
//
// Run: go test -tags=integration -v ./integration/... -run Issue49
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue49_Detokenize(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	created, createResp, err := client.Identity.Create(blnkgo.Identity{
		IdentityType: blnkgo.Individual,
		FirstName:    "Detokenize",
		LastName:     "Multi",
		EmailAddress: fmt.Sprintf("issue49-%s@example.com", suffix),
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

	_, tokenizeResp, err := client.Identity.Tokenize(created.IdentityId, blnkgo.TokenizeRequest{
		Fields: []blnkgo.TokenizableIdentityField{
			blnkgo.TokenizableFieldFirstName,
			blnkgo.TokenizableFieldLastName,
		},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, tokenizeResp.StatusCode)

	detokenized, detokenizeResp, err := client.Identity.Detokenize(created.IdentityId, blnkgo.DetokenizeRequest{
		Fields: []blnkgo.TokenizableIdentityField{
			blnkgo.TokenizableFieldFirstName,
			blnkgo.TokenizableFieldLastName,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, detokenizeResp)
	require.Equal(t, http.StatusOK, detokenizeResp.StatusCode)
	require.Equal(t, "Detokenize", detokenized.Fields["FirstName"])
	require.Equal(t, "Multi", detokenized.Fields["LastName"])
}

func TestIssue49_Detokenize_ValidationError(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Identity.Detokenize("", blnkgo.DetokenizeRequest{
		Fields: []blnkgo.TokenizableIdentityField{blnkgo.TokenizableFieldFirstName},
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "identity id is required")
}

func TestIssue49_Detokenize_NilFields(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Identity.Detokenize("idt_test_123", blnkgo.DetokenizeRequest{})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "fields must be an array")
}
