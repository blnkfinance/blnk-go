package blnkgo

import (
	"strings"
	"testing"
)

var knownErrorCodePrefixes = []string{
	"GEN", "AUTH", "APIKEY", "TXN", "BAL", "LGR", "ACC", "IDT", "RECON",
	"META", "HOOK", "QUEUE", "SRCH", "ADMIN",
}

func TestErrorCodeCatalogue_CountUniqueAndPrefixed(t *testing.T) {
	seen := make(map[string]struct{}, len(errorCodeCatalogue))
	prefixSet := make(map[string]struct{}, len(knownErrorCodePrefixes))
	for _, p := range knownErrorCodePrefixes {
		prefixSet[p] = struct{}{}
	}

	if got := len(errorCodeCatalogue); got != 79 {
		t.Fatalf("expected the full Core 0.15.4 catalogue (79 codes), got %d", got)
	}

	for _, code := range errorCodeCatalogue {
		if code == "" {
			t.Fatal("empty error code in catalogue")
		}
		if _, dup := seen[code]; dup {
			t.Fatalf("duplicate error code %q", code)
		}
		seen[code] = struct{}{}

		idx := strings.IndexByte(code, '_')
		if idx <= 0 {
			t.Fatalf("error code %q has no Core prefix", code)
		}
		prefix := code[:idx]
		if _, ok := prefixSet[prefix]; !ok {
			t.Fatalf("unknown Core prefix %q on %q", prefix, code)
		}
	}
}

func TestErrorCodeConstants_MatchExpectedStrings(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"ErrorCodeGenMalformedRequest", ErrorCodeGenMalformedRequest, "GEN_MALFORMED_REQUEST"},
		{"ErrorCodeGenValidationError", ErrorCodeGenValidationError, "GEN_VALIDATION_ERROR"},
		{"ErrorCodeGenMissingParameter", ErrorCodeGenMissingParameter, "GEN_MISSING_PARAMETER"},
		{"ErrorCodeGenBadRequest", ErrorCodeGenBadRequest, "GEN_BAD_REQUEST"},
		{"ErrorCodeGenNotFound", ErrorCodeGenNotFound, "GEN_NOT_FOUND"},
		{"ErrorCodeGenConflict", ErrorCodeGenConflict, "GEN_CONFLICT"},
		{"ErrorCodeGenResourceLocked", ErrorCodeGenResourceLocked, "GEN_RESOURCE_LOCKED"},
		{"ErrorCodeGenPayloadTooLarge", ErrorCodeGenPayloadTooLarge, "GEN_PAYLOAD_TOO_LARGE"},
		{"ErrorCodeGenRateLimited", ErrorCodeGenRateLimited, "GEN_RATE_LIMITED"},
		{"ErrorCodeGenInternal", ErrorCodeGenInternal, "GEN_INTERNAL"},
		{"ErrorCodeAuthMissingAPIKey", ErrorCodeAuthMissingAPIKey, "AUTH_MISSING_API_KEY"},
		{"ErrorCodeAuthInvalidAPIKey", ErrorCodeAuthInvalidAPIKey, "AUTH_INVALID_API_KEY"},
		{"ErrorCodeAuthExpiredAPIKey", ErrorCodeAuthExpiredAPIKey, "AUTH_EXPIRED_API_KEY"},
		{"ErrorCodeAuthMissingPrincipal", ErrorCodeAuthMissingPrincipal, "AUTH_MISSING_PRINCIPAL"},
		{"ErrorCodeAuthInsufficientPermissions", ErrorCodeAuthInsufficientPermissions, "AUTH_INSUFFICIENT_PERMISSIONS"},
		{"ErrorCodeAuthUnknownResource", ErrorCodeAuthUnknownResource, "AUTH_UNKNOWN_RESOURCE"},
		{"ErrorCodeAuthMasterKeyRequired", ErrorCodeAuthMasterKeyRequired, "AUTH_MASTER_KEY_REQUIRED"},
		{"ErrorCodeAuthCrossOwnerAccess", ErrorCodeAuthCrossOwnerAccess, "AUTH_CROSS_OWNER_ACCESS"},
		{"ErrorCodeAuthScopeEscalation", ErrorCodeAuthScopeEscalation, "AUTH_SCOPE_ESCALATION"},
		{"ErrorCodeAuthMetricsTokenRequired", ErrorCodeAuthMetricsTokenRequired, "AUTH_METRICS_TOKEN_REQUIRED"},
		{"ErrorCodeAuthInvalidBearerToken", ErrorCodeAuthInvalidBearerToken, "AUTH_INVALID_BEARER_TOKEN"},
		{"ErrorCodeAuthMetricsDisabled", ErrorCodeAuthMetricsDisabled, "AUTH_METRICS_DISABLED"},
		{"ErrorCodeAPIKeyNotFound", ErrorCodeAPIKeyNotFound, "APIKEY_NOT_FOUND"},
		{"ErrorCodeAPIKeyOwnerRequired", ErrorCodeAPIKeyOwnerRequired, "APIKEY_OWNER_REQUIRED"},
		{"ErrorCodeAPIKeyInvalid", ErrorCodeAPIKeyInvalid, "APIKEY_INVALID"},
		{"ErrorCodeTxnNotFound", ErrorCodeTxnNotFound, "TXN_NOT_FOUND"},
		{"ErrorCodeTxnInsufficientFunds", ErrorCodeTxnInsufficientFunds, "TXN_INSUFFICIENT_FUNDS"},
		{"ErrorCodeTxnInvalidAmount", ErrorCodeTxnInvalidAmount, "TXN_INVALID_AMOUNT"},
		{"ErrorCodeTxnPrecisionNotInteger", ErrorCodeTxnPrecisionNotInteger, "TXN_PRECISION_NOT_INTEGER"},
		{"ErrorCodeTxnInvalidDistribution", ErrorCodeTxnInvalidDistribution, "TXN_INVALID_DISTRIBUTION"},
		{"ErrorCodeTxnDuplicateReference", ErrorCodeTxnDuplicateReference, "TXN_DUPLICATE_REFERENCE"},
		{"ErrorCodeTxnNotInflight", ErrorCodeTxnNotInflight, "TXN_NOT_INFLIGHT"},
		{"ErrorCodeTxnAlreadyCommitted", ErrorCodeTxnAlreadyCommitted, "TXN_ALREADY_COMMITTED"},
		{"ErrorCodeTxnAlreadyVoided", ErrorCodeTxnAlreadyVoided, "TXN_ALREADY_VOIDED"},
		{"ErrorCodeTxnAlreadyRefunded", ErrorCodeTxnAlreadyRefunded, "TXN_ALREADY_REFUNDED"},
		{"ErrorCodeTxnCommitAmountExceeded", ErrorCodeTxnCommitAmountExceeded, "TXN_COMMIT_AMOUNT_EXCEEDED"},
		{"ErrorCodeTxnInvalidStatusAction", ErrorCodeTxnInvalidStatusAction, "TXN_INVALID_STATUS_ACTION"},
		{"ErrorCodeTxnBulkEmpty", ErrorCodeTxnBulkEmpty, "TXN_BULK_EMPTY"},
		{"ErrorCodeTxnBulkLimitExceeded", ErrorCodeTxnBulkLimitExceeded, "TXN_BULK_LIMIT_EXCEEDED"},
		{"ErrorCodeTxnValidationError", ErrorCodeTxnValidationError, "TXN_VALIDATION_ERROR"},
		{"ErrorCodeBalNotFound", ErrorCodeBalNotFound, "BAL_NOT_FOUND"},
		{"ErrorCodeBalHistoryNotFound", ErrorCodeBalHistoryNotFound, "BAL_HISTORY_NOT_FOUND"},
		{"ErrorCodeBalInvalidTimestamp", ErrorCodeBalInvalidTimestamp, "BAL_INVALID_TIMESTAMP"},
		{"ErrorCodeBalValidationError", ErrorCodeBalValidationError, "BAL_VALIDATION_ERROR"},
		{"ErrorCodeBalMonitorNotFound", ErrorCodeBalMonitorNotFound, "BAL_MONITOR_NOT_FOUND"},
		{"ErrorCodeLgrNotFound", ErrorCodeLgrNotFound, "LGR_NOT_FOUND"},
		{"ErrorCodeLgrDuplicate", ErrorCodeLgrDuplicate, "LGR_DUPLICATE"},
		{"ErrorCodeAccNotFound", ErrorCodeAccNotFound, "ACC_NOT_FOUND"},
		{"ErrorCodeAccDuplicate", ErrorCodeAccDuplicate, "ACC_DUPLICATE"},
		{"ErrorCodeAccGenerationFailed", ErrorCodeAccGenerationFailed, "ACC_GENERATION_FAILED"},
		{"ErrorCodeIdtNotFound", ErrorCodeIdtNotFound, "IDT_NOT_FOUND"},
		{"ErrorCodeIdtValidationError", ErrorCodeIdtValidationError, "IDT_VALIDATION_ERROR"},
		{"ErrorCodeIdtFieldNotTokenizable", ErrorCodeIdtFieldNotTokenizable, "IDT_FIELD_NOT_TOKENIZABLE"},
		{"ErrorCodeIdtFieldAlreadyTokenized", ErrorCodeIdtFieldAlreadyTokenized, "IDT_FIELD_ALREADY_TOKENIZED"},
		{"ErrorCodeIdtFieldNotTokenized", ErrorCodeIdtFieldNotTokenized, "IDT_FIELD_NOT_TOKENIZED"},
		{"ErrorCodeIdtFieldNotFound", ErrorCodeIdtFieldNotFound, "IDT_FIELD_NOT_FOUND"},
		{"ErrorCodeIdtTokenizationDisabled", ErrorCodeIdtTokenizationDisabled, "IDT_TOKENIZATION_DISABLED"},
		{"ErrorCodeReconNotFound", ErrorCodeReconNotFound, "RECON_NOT_FOUND"},
		{"ErrorCodeReconRuleNotFound", ErrorCodeReconRuleNotFound, "RECON_RULE_NOT_FOUND"},
		{"ErrorCodeReconUploadFailed", ErrorCodeReconUploadFailed, "RECON_UPLOAD_FAILED"},
		{"ErrorCodeReconUploadProcessingFailed", ErrorCodeReconUploadProcessingFailed, "RECON_UPLOAD_PROCESSING_FAILED"},
		{"ErrorCodeReconUploadURLInvalid", ErrorCodeReconUploadURLInvalid, "RECON_UPLOAD_URL_INVALID"},
		{"ErrorCodeReconUploadHostNotAllowed", ErrorCodeReconUploadHostNotAllowed, "RECON_UPLOAD_HOST_NOT_ALLOWED"},
		{"ErrorCodeReconRuleInvalid", ErrorCodeReconRuleInvalid, "RECON_RULE_INVALID"},
		{"ErrorCodeReconMatchingRulesRequired", ErrorCodeReconMatchingRulesRequired, "RECON_MATCHING_RULES_REQUIRED"},
		{"ErrorCodeReconExternalTxnsRequired", ErrorCodeReconExternalTxnsRequired, "RECON_EXTERNAL_TXNS_REQUIRED"},
		{"ErrorCodeReconStartFailed", ErrorCodeReconStartFailed, "RECON_START_FAILED"},
		{"ErrorCodeMetaEntityNotFound", ErrorCodeMetaEntityNotFound, "META_ENTITY_NOT_FOUND"},
		{"ErrorCodeMetaUnsupportedEntity", ErrorCodeMetaUnsupportedEntity, "META_UNSUPPORTED_ENTITY"},
		{"ErrorCodeMetaInvalidEntityID", ErrorCodeMetaInvalidEntityID, "META_INVALID_ENTITY_ID"},
		{"ErrorCodeHookNotFound", ErrorCodeHookNotFound, "HOOK_NOT_FOUND"},
		{"ErrorCodeHookInvalid", ErrorCodeHookInvalid, "HOOK_INVALID"},
		{"ErrorCodeHookOperationFailed", ErrorCodeHookOperationFailed, "HOOK_OPERATION_FAILED"},
		{"ErrorCodeQueueBackpressure", ErrorCodeQueueBackpressure, "QUEUE_BACKPRESSURE"},
		{"ErrorCodeSrchQueryInvalid", ErrorCodeSrchQueryInvalid, "SRCH_QUERY_INVALID"},
		{"ErrorCodeSrchFailed", ErrorCodeSrchFailed, "SRCH_FAILED"},
		{"ErrorCodeSrchReindexInProgress", ErrorCodeSrchReindexInProgress, "SRCH_REINDEX_IN_PROGRESS"},
		{"ErrorCodeSrchReindexNotStarted", ErrorCodeSrchReindexNotStarted, "SRCH_REINDEX_NOT_STARTED"},
		{"ErrorCodeAdminBackupFailed", ErrorCodeAdminBackupFailed, "ADMIN_BACKUP_FAILED"},
	}

	if len(cases) != 79 {
		t.Fatalf("expected 79 constant assertions, got %d", len(cases))
	}

	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, tc.got, tc.want)
		}
	}
}

func TestErrorCodeConstants_Core0154CodesPresent(t *testing.T) {
	if ErrorCodeTxnAlreadyRefunded != "TXN_ALREADY_REFUNDED" {
		t.Fatalf("ErrorCodeTxnAlreadyRefunded = %q", ErrorCodeTxnAlreadyRefunded)
	}
	if ErrorCodeBalNotFound != "BAL_NOT_FOUND" {
		t.Fatalf("ErrorCodeBalNotFound = %q", ErrorCodeBalNotFound)
	}
	if ErrorCodeTxnValidationError != "TXN_VALIDATION_ERROR" {
		t.Fatalf("ErrorCodeTxnValidationError = %q", ErrorCodeTxnValidationError)
	}
	if ErrorCodeGenConflict != "GEN_CONFLICT" {
		t.Fatalf("ErrorCodeGenConflict = %q", ErrorCodeGenConflict)
	}
}

func TestErrorCodeConstants_LegacyUnchanged(t *testing.T) {
	if ErrorCodeTxnInvalidAmount != "TXN_INVALID_AMOUNT" {
		t.Fatalf("ErrorCodeTxnInvalidAmount changed: %q", ErrorCodeTxnInvalidAmount)
	}
	if ErrorCodeGenConflict != "GEN_CONFLICT" {
		t.Fatalf("ErrorCodeGenConflict changed: %q", ErrorCodeGenConflict)
	}
	if ErrorCodeTxnValidationError != "TXN_VALIDATION_ERROR" {
		t.Fatalf("ErrorCodeTxnValidationError changed: %q", ErrorCodeTxnValidationError)
	}
}
