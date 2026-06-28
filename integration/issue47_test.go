//go:build integration

// Integration tests for issue #47 — Identity.GetTokenizedFields (GET /identities/{id}/tokenized-fields).
// Requires Blnk Core running at http://localhost:5001 with tokenization enabled.
//
// Run: go test -tags=integration -v ./integration/... -run Issue47
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue47_GetTokenizedFields_Empty(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	created, createResp, err := client.Identity.Create(blnkgo.Identity{
		IdentityType: blnkgo.Individual,
		FirstName:    "Fields",
		LastName:     "Empty",
		EmailAddress: fmt.Sprintf("issue47-empty-%s@example.com", suffix),
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

	fields, resp, err := client.Identity.GetTokenizedFields(created.IdentityId)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Empty(t, fields.TokenizedFields)
}

func TestIssue47_GetTokenizedFields_AfterTokenize(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	created, createResp, err := client.Identity.Create(blnkgo.Identity{
		IdentityType: blnkgo.Individual,
		FirstName:    "Fields",
		LastName:     "Tokenized",
		EmailAddress: fmt.Sprintf("issue47-tokenized-%s@example.com", suffix),
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

	fields, resp, err := client.Identity.GetTokenizedFields(created.IdentityId)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, fields.TokenizedFields, 2)

	fieldSet := map[blnkgo.TokenizableIdentityField]bool{}
	for _, field := range fields.TokenizedFields {
		fieldSet[field] = true
	}
	require.True(t, fieldSet[blnkgo.TokenizableFieldFirstName])
	require.True(t, fieldSet[blnkgo.TokenizableFieldLastName])
}

func TestIssue47_GetTokenizedFields_ValidationError(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Identity.GetTokenizedFields("")
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "identity id is required")
}

func TestIssue47_GetTokenizedFields_NotFound(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Identity.GetTokenizedFields("idt_nonexistent_issue47")
	require.Error(t, err)
	require.NotNil(t, resp)
	require.NotEqual(t, http.StatusOK, resp.StatusCode)
}
