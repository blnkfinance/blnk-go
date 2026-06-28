package blnkgo

import (
	"fmt"
	"regexp"
	"strings"
)

var identityIDUUIDSuffix = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func validateIdentityID(id string) error {
	if id == "" {
		return nil
	}
	if !strings.HasPrefix(id, "idt_") {
		return fmt.Errorf("identity_id must start with the 'idt_' prefix")
	}
	suffix := strings.TrimPrefix(id, "idt_")
	if !identityIDUUIDSuffix.MatchString(suffix) {
		return fmt.Errorf("identity_id suffix after 'idt_' must be a valid UUID")
	}
	return nil
}

// ValidateCreateIdentity performs client-side checks before POST /identities.
// Field requirements match the Blnk API: optional fields are not enforced here.
func ValidateCreateIdentity(identity Identity) error {
	if identity.IdentityType != "" && identity.IdentityType != Individual && identity.IdentityType != Organization {
		return fmt.Errorf("invalid identity_type: must be individual or organization")
	}
	return validateIdentityID(identity.IdentityID)
}

// ValidateIdentityID performs client-side checks before identity operations that require an ID.
func ValidateIdentityID(identityID string) error {
	if strings.TrimSpace(identityID) == "" {
		return fmt.Errorf("identity id is required")
	}
	return nil
}

// ValidateTokenizeIdentityField performs client-side checks before POST /identities/{id}/tokenize/{field}.
func ValidateTokenizeIdentityField(identityID, field string) error {
	if err := ValidateIdentityID(identityID); err != nil {
		return err
	}
	if strings.TrimSpace(field) == "" {
		return fmt.Errorf("field name is required")
	}
	return nil
}

// ValidateTokenizeIdentityRequest performs client-side checks before POST /identities/{id}/tokenize.
func ValidateTokenizeIdentityRequest(identityID string, body TokenizeRequest) error {
	if err := ValidateIdentityID(identityID); err != nil {
		return err
	}
	if len(body.Fields) == 0 {
		return fmt.Errorf("at least one field must be specified")
	}
	for _, field := range body.Fields {
		if strings.TrimSpace(string(field)) == "" {
			return fmt.Errorf("each field must be a non-empty string")
		}
	}
	return nil
}

// ValidateDetokenizeIdentityRequest performs client-side checks before POST /identities/{id}/detokenize.
func ValidateDetokenizeIdentityRequest(identityID string, body DetokenizeRequest) error {
	if err := ValidateIdentityID(identityID); err != nil {
		return err
	}
	if body.Fields == nil {
		return fmt.Errorf("fields must be an array")
	}
	for _, field := range body.Fields {
		if strings.TrimSpace(string(field)) == "" {
			return fmt.Errorf("each field must be a non-empty string")
		}
	}
	return nil
}
