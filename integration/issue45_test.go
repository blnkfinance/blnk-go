//go:build integration

// Integration tests for issue #45 — Identity.TokenizeField (POST /identities/{id}/tokenize/{field}).
// Requires Blnk Core running at http://localhost:5001 with tokenization enabled.
//
// Run: go test -tags=integration -v ./integration/... -run Issue45
package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestIssue45_TokenizeField(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	dob := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)

	created, createResp, err := client.Identity.Create(blnkgo.Identity{
		IdentityType: blnkgo.Individual,
		FirstName:    "Tokenize",
		LastName:     "Field",
		EmailAddress: fmt.Sprintf("issue45-%s@example.com", suffix),
		PhoneNumber:  "1234567890",
		Category:     "customer",
		Street:       "123 Main St",
		Country:      "USA",
		State:        "CA",
		PostCode:     "90001",
		City:         "Los Angeles",
		DOB:          &dob,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	require.NotEmpty(t, created.IdentityId)

	tokenized, tokenizeResp, err := client.Identity.TokenizeField(created.IdentityId, string(blnkgo.TokenizableFieldFirstName))
	require.NoError(t, err)
	require.NotNil(t, tokenizeResp)
	require.Equal(t, http.StatusOK, tokenizeResp.StatusCode)
	require.Equal(t, "Field tokenized successfully", tokenized.Message)
}

func TestIssue45_TokenizeField_ValidationError(t *testing.T) {
	client := newIntegrationClient(t)

	_, resp, err := client.Identity.TokenizeField("", string(blnkgo.TokenizableFieldFirstName))
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "identity id is required")
}

func TestIssue45_TokenizeField_AlreadyTokenized(t *testing.T) {
	client := newIntegrationClient(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	created, createResp, err := client.Identity.Create(blnkgo.Identity{
		IdentityType: blnkgo.Individual,
		FirstName:    "Already",
		LastName:     "Tokenized",
		EmailAddress: fmt.Sprintf("issue45-dup-%s@example.com", suffix),
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

	_, firstResp, err := client.Identity.TokenizeField(created.IdentityId, string(blnkgo.TokenizableFieldLastName))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, firstResp.StatusCode)

	_, secondResp, err := client.Identity.TokenizeField(created.IdentityId, string(blnkgo.TokenizableFieldLastName))
	require.Error(t, err)
	require.NotNil(t, secondResp)
	require.NotEqual(t, http.StatusOK, secondResp.StatusCode)
}
