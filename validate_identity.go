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
