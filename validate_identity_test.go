package blnkgo_test

import (
	"encoding/json"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestValidateCreateIdentity_MinimalIndividual(t *testing.T) {
	require.NoError(t, blnkgo.ValidateCreateIdentity(blnkgo.Identity{
		IdentityType: blnkgo.Individual,
		FirstName:    "Jane",
		Category:     "customer",
	}))
}

func TestValidateCreateIdentity_MinimalOrganization(t *testing.T) {
	require.NoError(t, blnkgo.ValidateCreateIdentity(blnkgo.Identity{
		IdentityType:     blnkgo.Organization,
		OrganizationName: "ACME Inc",
	}))
}

func TestValidateCreateIdentity_CallerSuppliedIdentityID(t *testing.T) {
	require.NoError(t, blnkgo.ValidateCreateIdentity(blnkgo.Identity{
		IdentityID:   "idt_8c5a8e2f-3f1d-5a9b-9c3e-4d8f1e5a7b2c",
		IdentityType: blnkgo.Individual,
		FirstName:    "Caller",
	}))
}

func TestValidateCreateIdentity_InvalidIdentityType(t *testing.T) {
	err := blnkgo.ValidateCreateIdentity(blnkgo.Identity{
		IdentityType: blnkgo.IdentityType("business"),
		FirstName:    "Jane",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid identity_type")
}

func TestValidateCreateIdentity_InvalidIdentityIDPrefix(t *testing.T) {
	err := blnkgo.ValidateCreateIdentity(blnkgo.Identity{
		IdentityID: "user_8c5a8e2f-3f1d-5a9b-9c3e-4d8f1e5a7b2c",
		FirstName:  "Jane",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "idt_")
}

func TestValidateCreateIdentity_InvalidIdentityIDSuffix(t *testing.T) {
	err := blnkgo.ValidateCreateIdentity(blnkgo.Identity{
		IdentityID: "idt_not-a-uuid",
		FirstName:  "Jane",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "valid UUID")
}

func TestIdentity_JSONMarshal_CreateIncludesIdentityID(t *testing.T) {
	body := blnkgo.Identity{
		IdentityID:   "idt_8c5a8e2f-3f1d-5a9b-9c3e-4d8f1e5a7b2c",
		IdentityType: blnkgo.Individual,
		FirstName:    "Jane",
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(payload, &decoded))
	require.Equal(t, "idt_8c5a8e2f-3f1d-5a9b-9c3e-4d8f1e5a7b2c", decoded["identity_id"])
	require.Equal(t, "individual", decoded["identity_type"])
}

func TestIdentity_Update_JSONMarshal_IncludesEmptyStrings(t *testing.T) {
	body := blnkgo.Identity{
		FirstName:    "Jane",
		EmailAddress: "",
		PhoneNumber:  "",
		Category:     "",
		Street:       "",
		Country:      "",
		State:        "",
		PostCode:     "",
		City:         "",
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(payload, &decoded))
	require.Equal(t, "", decoded["email_address"])
	require.Equal(t, "", decoded["phone_number"])
	require.Equal(t, "", decoded["category"])
	require.Equal(t, "", decoded["street"])
	require.Equal(t, "", decoded["country"])
	require.Equal(t, "", decoded["state"])
	require.Equal(t, "", decoded["post_code"])
	require.Equal(t, "", decoded["city"])
}

func TestValidateIdentityID(t *testing.T) {
	require.NoError(t, blnkgo.ValidateIdentityID("idt_test_123"))
	require.Error(t, blnkgo.ValidateIdentityID(""))
	require.Error(t, blnkgo.ValidateIdentityID("   "))
}

func TestValidateTokenizeIdentityField(t *testing.T) {
	require.NoError(t, blnkgo.ValidateTokenizeIdentityField("idt_test_123", "FirstName"))
	require.Error(t, blnkgo.ValidateTokenizeIdentityField("", "FirstName"))
	require.Error(t, blnkgo.ValidateTokenizeIdentityField("idt_test_123", ""))
}

func TestValidateTokenizeIdentityRequest(t *testing.T) {
	body := blnkgo.TokenizeRequest{
		Fields: []blnkgo.TokenizableIdentityField{
			blnkgo.TokenizableFieldFirstName,
			blnkgo.TokenizableFieldEmailAddress,
		},
	}
	require.NoError(t, blnkgo.ValidateTokenizeIdentityRequest("idt_test_123", body))
	require.Error(t, blnkgo.ValidateTokenizeIdentityRequest("", body))
	require.Error(t, blnkgo.ValidateTokenizeIdentityRequest("idt_test_123", blnkgo.TokenizeRequest{}))
	require.Error(t, blnkgo.ValidateTokenizeIdentityRequest("idt_test_123", blnkgo.TokenizeRequest{
		Fields: []blnkgo.TokenizableIdentityField{""},
	}))
}

func TestValidateDetokenizeIdentityRequest(t *testing.T) {
	body := blnkgo.DetokenizeRequest{
		Fields: []blnkgo.TokenizableIdentityField{
			blnkgo.TokenizableFieldFirstName,
			blnkgo.TokenizableFieldEmailAddress,
		},
	}
	require.NoError(t, blnkgo.ValidateDetokenizeIdentityRequest("idt_test_123", body))
	require.NoError(t, blnkgo.ValidateDetokenizeIdentityRequest("idt_test_123", blnkgo.DetokenizeRequest{Fields: []blnkgo.TokenizableIdentityField{}}))
	require.Error(t, blnkgo.ValidateDetokenizeIdentityRequest("", body))
	require.Error(t, blnkgo.ValidateDetokenizeIdentityRequest("idt_test_123", blnkgo.DetokenizeRequest{}))
	require.Error(t, blnkgo.ValidateDetokenizeIdentityRequest("idt_test_123", blnkgo.DetokenizeRequest{
		Fields: []blnkgo.TokenizableIdentityField{""},
	}))
}
