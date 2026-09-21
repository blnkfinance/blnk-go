package blnkgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Core error_detail.code values, mirroring internal/apierror/codes.go in Blnk
// Core 0.15.4. Compare ErrorDetail.Code — do not branch on message text.
// Each comment gives the HTTP status Core sends with the code.
//
// https://docs.blnkfinance.com/advanced/error-codes
const (
	// GEN

	// ErrorCodeGenMalformedRequest is 400 when the body could not be decoded.
	ErrorCodeGenMalformedRequest = "GEN_MALFORMED_REQUEST"
	// ErrorCodeGenValidationError is 400.
	ErrorCodeGenValidationError = "GEN_VALIDATION_ERROR"
	// ErrorCodeGenMissingParameter is 400.
	ErrorCodeGenMissingParameter = "GEN_MISSING_PARAMETER"
	// ErrorCodeGenBadRequest is 400.
	ErrorCodeGenBadRequest = "GEN_BAD_REQUEST"
	// ErrorCodeGenNotFound is 404.
	ErrorCodeGenNotFound = "GEN_NOT_FOUND"
	// ErrorCodeGenConflict is 409 for a duplicate GL indicator + currency, or a
	// multi-leg refund that failed part-way (0.15.4+).
	ErrorCodeGenConflict = "GEN_CONFLICT"
	// ErrorCodeGenResourceLocked is 423.
	ErrorCodeGenResourceLocked = "GEN_RESOURCE_LOCKED"
	// ErrorCodeGenPayloadTooLarge is 413.
	ErrorCodeGenPayloadTooLarge = "GEN_PAYLOAD_TOO_LARGE"
	// ErrorCodeGenRateLimited is 429.
	ErrorCodeGenRateLimited = "GEN_RATE_LIMITED"
	// ErrorCodeGenInternal is 500.
	ErrorCodeGenInternal = "GEN_INTERNAL"

	// AUTH

	// ErrorCodeAuthMissingAPIKey is 401.
	ErrorCodeAuthMissingAPIKey = "AUTH_MISSING_API_KEY"
	// ErrorCodeAuthInvalidAPIKey is 401.
	ErrorCodeAuthInvalidAPIKey = "AUTH_INVALID_API_KEY"
	// ErrorCodeAuthExpiredAPIKey is 401.
	ErrorCodeAuthExpiredAPIKey = "AUTH_EXPIRED_API_KEY"
	// ErrorCodeAuthMissingPrincipal is 401.
	ErrorCodeAuthMissingPrincipal = "AUTH_MISSING_PRINCIPAL"
	// ErrorCodeAuthInsufficientPermissions is 403.
	ErrorCodeAuthInsufficientPermissions = "AUTH_INSUFFICIENT_PERMISSIONS"
	// ErrorCodeAuthUnknownResource is 403.
	ErrorCodeAuthUnknownResource = "AUTH_UNKNOWN_RESOURCE"
	// ErrorCodeAuthMasterKeyRequired is 403.
	ErrorCodeAuthMasterKeyRequired = "AUTH_MASTER_KEY_REQUIRED"
	// ErrorCodeAuthCrossOwnerAccess is 403.
	ErrorCodeAuthCrossOwnerAccess = "AUTH_CROSS_OWNER_ACCESS"
	// ErrorCodeAuthScopeEscalation is 403 when a key is granted scopes it does not hold.
	ErrorCodeAuthScopeEscalation = "AUTH_SCOPE_ESCALATION"
	// ErrorCodeAuthMetricsTokenRequired is 401.
	ErrorCodeAuthMetricsTokenRequired = "AUTH_METRICS_TOKEN_REQUIRED"
	// ErrorCodeAuthInvalidBearerToken is 401.
	ErrorCodeAuthInvalidBearerToken = "AUTH_INVALID_BEARER_TOKEN"
	// ErrorCodeAuthMetricsDisabled is 403.
	ErrorCodeAuthMetricsDisabled = "AUTH_METRICS_DISABLED"

	// APIKEY

	// ErrorCodeAPIKeyNotFound is 404.
	ErrorCodeAPIKeyNotFound = "APIKEY_NOT_FOUND"
	// ErrorCodeAPIKeyOwnerRequired is 400.
	ErrorCodeAPIKeyOwnerRequired = "APIKEY_OWNER_REQUIRED"
	// ErrorCodeAPIKeyInvalid is 400.
	ErrorCodeAPIKeyInvalid = "APIKEY_INVALID"

	// TXN

	// ErrorCodeTxnNotFound is 404 when looking up by id or reference.
	ErrorCodeTxnNotFound = "TXN_NOT_FOUND"
	// ErrorCodeTxnInsufficientFunds is 400 when overdraft is off.
	ErrorCodeTxnInsufficientFunds = "TXN_INSUFFICIENT_FUNDS"
	// ErrorCodeTxnInvalidAmount is 400. Some Core versions report this as
	// ErrorCodeTxnValidationError; match either.
	ErrorCodeTxnInvalidAmount = "TXN_INVALID_AMOUNT"
	// ErrorCodeTxnPrecisionNotInteger is 400.
	ErrorCodeTxnPrecisionNotInteger = "TXN_PRECISION_NOT_INTEGER"
	// ErrorCodeTxnInvalidDistribution is 400 when multi-source/destination legs do not add up.
	ErrorCodeTxnInvalidDistribution = "TXN_INVALID_DISTRIBUTION"
	// ErrorCodeTxnDuplicateReference is 409.
	ErrorCodeTxnDuplicateReference = "TXN_DUPLICATE_REFERENCE"
	// ErrorCodeTxnNotInflight is 400.
	ErrorCodeTxnNotInflight = "TXN_NOT_INFLIGHT"
	// ErrorCodeTxnAlreadyCommitted is 409.
	ErrorCodeTxnAlreadyCommitted = "TXN_ALREADY_COMMITTED"
	// ErrorCodeTxnAlreadyVoided is 409.
	ErrorCodeTxnAlreadyVoided = "TXN_ALREADY_VOIDED"
	// ErrorCodeTxnAlreadyRefunded is 409 when refunding twice or refunding a refund (0.15.4+).
	ErrorCodeTxnAlreadyRefunded = "TXN_ALREADY_REFUNDED"
	// ErrorCodeTxnCommitAmountExceeded is 400.
	ErrorCodeTxnCommitAmountExceeded = "TXN_COMMIT_AMOUNT_EXCEEDED"
	// ErrorCodeTxnInvalidStatusAction is 400 when status is not commit or void.
	ErrorCodeTxnInvalidStatusAction = "TXN_INVALID_STATUS_ACTION"
	// ErrorCodeTxnBulkEmpty is 400.
	ErrorCodeTxnBulkEmpty = "TXN_BULK_EMPTY"
	// ErrorCodeTxnBulkLimitExceeded is 400.
	ErrorCodeTxnBulkLimitExceeded = "TXN_BULK_LIMIT_EXCEEDED"
	// ErrorCodeTxnValidationError is 400, including a split request with both
	// sources and destinations (0.15.4+).
	ErrorCodeTxnValidationError = "TXN_VALIDATION_ERROR"

	// BAL

	// ErrorCodeBalNotFound is 404. Since 0.15.4 a transaction naming a missing
	// balance returns this, not ErrorCodeTxnNotFound.
	ErrorCodeBalNotFound = "BAL_NOT_FOUND"
	// ErrorCodeBalHistoryNotFound is 404 when there is no history at the requested timestamp.
	ErrorCodeBalHistoryNotFound = "BAL_HISTORY_NOT_FOUND"
	// ErrorCodeBalInvalidTimestamp is 400.
	ErrorCodeBalInvalidTimestamp = "BAL_INVALID_TIMESTAMP"
	// ErrorCodeBalValidationError is 400.
	ErrorCodeBalValidationError = "BAL_VALIDATION_ERROR"
	// ErrorCodeBalMonitorNotFound is 404.
	ErrorCodeBalMonitorNotFound = "BAL_MONITOR_NOT_FOUND"

	// LGR

	// ErrorCodeLgrNotFound is 404.
	ErrorCodeLgrNotFound = "LGR_NOT_FOUND"
	// ErrorCodeLgrDuplicate is 409.
	ErrorCodeLgrDuplicate = "LGR_DUPLICATE"

	// ACC

	// ErrorCodeAccNotFound is 404.
	ErrorCodeAccNotFound = "ACC_NOT_FOUND"
	// ErrorCodeAccDuplicate is 409.
	ErrorCodeAccDuplicate = "ACC_DUPLICATE"
	// ErrorCodeAccGenerationFailed is 500.
	ErrorCodeAccGenerationFailed = "ACC_GENERATION_FAILED"

	// IDT

	// ErrorCodeIdtNotFound is 404.
	ErrorCodeIdtNotFound = "IDT_NOT_FOUND"
	// ErrorCodeIdtValidationError is 400.
	ErrorCodeIdtValidationError = "IDT_VALIDATION_ERROR"
	// ErrorCodeIdtFieldNotTokenizable is 400.
	ErrorCodeIdtFieldNotTokenizable = "IDT_FIELD_NOT_TOKENIZABLE"
	// ErrorCodeIdtFieldAlreadyTokenized is 409.
	ErrorCodeIdtFieldAlreadyTokenized = "IDT_FIELD_ALREADY_TOKENIZED"
	// ErrorCodeIdtFieldNotTokenized is 400.
	ErrorCodeIdtFieldNotTokenized = "IDT_FIELD_NOT_TOKENIZED"
	// ErrorCodeIdtFieldNotFound is 400.
	ErrorCodeIdtFieldNotFound = "IDT_FIELD_NOT_FOUND"
	// ErrorCodeIdtTokenizationDisabled is 403.
	ErrorCodeIdtTokenizationDisabled = "IDT_TOKENIZATION_DISABLED"

	// RECON

	// ErrorCodeReconNotFound is 404.
	ErrorCodeReconNotFound = "RECON_NOT_FOUND"
	// ErrorCodeReconRuleNotFound is 404.
	ErrorCodeReconRuleNotFound = "RECON_RULE_NOT_FOUND"
	// ErrorCodeReconUploadFailed is 400.
	ErrorCodeReconUploadFailed = "RECON_UPLOAD_FAILED"
	// ErrorCodeReconUploadProcessingFailed is 500.
	ErrorCodeReconUploadProcessingFailed = "RECON_UPLOAD_PROCESSING_FAILED"
	// ErrorCodeReconUploadURLInvalid is 400.
	ErrorCodeReconUploadURLInvalid = "RECON_UPLOAD_URL_INVALID"
	// ErrorCodeReconUploadHostNotAllowed is 400.
	ErrorCodeReconUploadHostNotAllowed = "RECON_UPLOAD_HOST_NOT_ALLOWED"
	// ErrorCodeReconRuleInvalid is 400.
	ErrorCodeReconRuleInvalid = "RECON_RULE_INVALID"
	// ErrorCodeReconMatchingRulesRequired is 400.
	ErrorCodeReconMatchingRulesRequired = "RECON_MATCHING_RULES_REQUIRED"
	// ErrorCodeReconExternalTxnsRequired is 400.
	ErrorCodeReconExternalTxnsRequired = "RECON_EXTERNAL_TXNS_REQUIRED"
	// ErrorCodeReconStartFailed is 500.
	ErrorCodeReconStartFailed = "RECON_START_FAILED"

	// META

	// ErrorCodeMetaEntityNotFound is 404.
	ErrorCodeMetaEntityNotFound = "META_ENTITY_NOT_FOUND"
	// ErrorCodeMetaUnsupportedEntity is 400.
	ErrorCodeMetaUnsupportedEntity = "META_UNSUPPORTED_ENTITY"
	// ErrorCodeMetaInvalidEntityID is 400.
	ErrorCodeMetaInvalidEntityID = "META_INVALID_ENTITY_ID"

	// HOOK

	// ErrorCodeHookNotFound is 404.
	ErrorCodeHookNotFound = "HOOK_NOT_FOUND"
	// ErrorCodeHookInvalid is 400.
	ErrorCodeHookInvalid = "HOOK_INVALID"
	// ErrorCodeHookOperationFailed is 500.
	ErrorCodeHookOperationFailed = "HOOK_OPERATION_FAILED"

	// QUEUE

	// ErrorCodeQueueBackpressure is 503; retry later.
	ErrorCodeQueueBackpressure = "QUEUE_BACKPRESSURE"

	// SRCH

	// ErrorCodeSrchQueryInvalid is 400.
	ErrorCodeSrchQueryInvalid = "SRCH_QUERY_INVALID"
	// ErrorCodeSrchFailed is 500.
	ErrorCodeSrchFailed = "SRCH_FAILED"
	// ErrorCodeSrchReindexInProgress is 409.
	ErrorCodeSrchReindexInProgress = "SRCH_REINDEX_IN_PROGRESS"
	// ErrorCodeSrchReindexNotStarted is 404.
	ErrorCodeSrchReindexNotStarted = "SRCH_REINDEX_NOT_STARTED"

	// ADMIN

	// ErrorCodeAdminBackupFailed is 500.
	ErrorCodeAdminBackupFailed = "ADMIN_BACKUP_FAILED"
)

// errorCodeCatalogue is every Core 0.15.4 error_detail.code mirrored above.
// Kept for tests so the exported count stays aligned with Core.
var errorCodeCatalogue = []string{
	ErrorCodeGenMalformedRequest,
	ErrorCodeGenValidationError,
	ErrorCodeGenMissingParameter,
	ErrorCodeGenBadRequest,
	ErrorCodeGenNotFound,
	ErrorCodeGenConflict,
	ErrorCodeGenResourceLocked,
	ErrorCodeGenPayloadTooLarge,
	ErrorCodeGenRateLimited,
	ErrorCodeGenInternal,
	ErrorCodeAuthMissingAPIKey,
	ErrorCodeAuthInvalidAPIKey,
	ErrorCodeAuthExpiredAPIKey,
	ErrorCodeAuthMissingPrincipal,
	ErrorCodeAuthInsufficientPermissions,
	ErrorCodeAuthUnknownResource,
	ErrorCodeAuthMasterKeyRequired,
	ErrorCodeAuthCrossOwnerAccess,
	ErrorCodeAuthScopeEscalation,
	ErrorCodeAuthMetricsTokenRequired,
	ErrorCodeAuthInvalidBearerToken,
	ErrorCodeAuthMetricsDisabled,
	ErrorCodeAPIKeyNotFound,
	ErrorCodeAPIKeyOwnerRequired,
	ErrorCodeAPIKeyInvalid,
	ErrorCodeTxnNotFound,
	ErrorCodeTxnInsufficientFunds,
	ErrorCodeTxnInvalidAmount,
	ErrorCodeTxnPrecisionNotInteger,
	ErrorCodeTxnInvalidDistribution,
	ErrorCodeTxnDuplicateReference,
	ErrorCodeTxnNotInflight,
	ErrorCodeTxnAlreadyCommitted,
	ErrorCodeTxnAlreadyVoided,
	ErrorCodeTxnAlreadyRefunded,
	ErrorCodeTxnCommitAmountExceeded,
	ErrorCodeTxnInvalidStatusAction,
	ErrorCodeTxnBulkEmpty,
	ErrorCodeTxnBulkLimitExceeded,
	ErrorCodeTxnValidationError,
	ErrorCodeBalNotFound,
	ErrorCodeBalHistoryNotFound,
	ErrorCodeBalInvalidTimestamp,
	ErrorCodeBalValidationError,
	ErrorCodeBalMonitorNotFound,
	ErrorCodeLgrNotFound,
	ErrorCodeLgrDuplicate,
	ErrorCodeAccNotFound,
	ErrorCodeAccDuplicate,
	ErrorCodeAccGenerationFailed,
	ErrorCodeIdtNotFound,
	ErrorCodeIdtValidationError,
	ErrorCodeIdtFieldNotTokenizable,
	ErrorCodeIdtFieldAlreadyTokenized,
	ErrorCodeIdtFieldNotTokenized,
	ErrorCodeIdtFieldNotFound,
	ErrorCodeIdtTokenizationDisabled,
	ErrorCodeReconNotFound,
	ErrorCodeReconRuleNotFound,
	ErrorCodeReconUploadFailed,
	ErrorCodeReconUploadProcessingFailed,
	ErrorCodeReconUploadURLInvalid,
	ErrorCodeReconUploadHostNotAllowed,
	ErrorCodeReconRuleInvalid,
	ErrorCodeReconMatchingRulesRequired,
	ErrorCodeReconExternalTxnsRequired,
	ErrorCodeReconStartFailed,
	ErrorCodeMetaEntityNotFound,
	ErrorCodeMetaUnsupportedEntity,
	ErrorCodeMetaInvalidEntityID,
	ErrorCodeHookNotFound,
	ErrorCodeHookInvalid,
	ErrorCodeHookOperationFailed,
	ErrorCodeQueueBackpressure,
	ErrorCodeSrchQueryInvalid,
	ErrorCodeSrchFailed,
	ErrorCodeSrchReindexInProgress,
	ErrorCodeSrchReindexNotStarted,
	ErrorCodeAdminBackupFailed,
}

// ApiErrorDetail is the structured error payload returned by Core 0.15.0+.
type ApiErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ApiErrorResponse represents a non-success HTTP response from the Blnk API.
type ApiErrorResponse struct {
	Status      int             `json:"status"`
	Message     string          `json:"message"`
	LegacyError string          `json:"error,omitempty"`
	ErrorDetail *ApiErrorDetail `json:"error_detail,omitempty"`
	Body        []byte          `json:"body"`
}

func (a *ApiErrorResponse) Error() string {
	if a.ErrorDetail != nil && a.ErrorDetail.Code != "" {
		return fmt.Sprintf("Status: %d, Code: %s, Message: %s", a.Status, a.ErrorDetail.Code, a.ErrorDetail.Message)
	}
	if a.ErrorDetail != nil && a.ErrorDetail.Message != "" {
		return fmt.Sprintf("Status: %d, Message: %s", a.Status, a.ErrorDetail.Message)
	}
	if a.LegacyError != "" {
		return fmt.Sprintf("Status: %d, Message: %s", a.Status, a.LegacyError)
	}
	return fmt.Sprintf("Status: %d, Message: %s, Body: %s", a.Status, a.Message, a.Body)
}

// AsApiErrorResponse returns the Blnk API error when err wraps ApiErrorResponse.
func AsApiErrorResponse(err error) (*ApiErrorResponse, bool) {
	var apiErr *ApiErrorResponse
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

// ParseApiErrorBody extracts structured error_detail (and legacy error) from a JSON body.
func ParseApiErrorBody(body []byte) (legacyError string, detail *ApiErrorDetail) {
	if len(body) == 0 {
		return "", nil
	}

	var payload struct {
		Error       string          `json:"error"`
		ErrorDetail *ApiErrorDetail `json:"error_detail"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", nil
	}

	if payload.ErrorDetail != nil && payload.ErrorDetail.Code != "" && payload.ErrorDetail.Message != "" {
		return payload.Error, payload.ErrorDetail
	}

	if payload.Error != "" {
		return payload.Error, &ApiErrorDetail{
			Code:    "UNKNOWN",
			Message: payload.Error,
		}
	}

	return payload.Error, nil
}

func newApiErrorResponse(statusCode int, statusText string, body []byte) *ApiErrorResponse {
	legacyError, detail := ParseApiErrorBody(body)
	return &ApiErrorResponse{
		Status:      statusCode,
		Message:     statusText,
		LegacyError: legacyError,
		ErrorDetail: detail,
		Body:        body,
	}
}

func (c *Client) CheckResponse(resp *http.Response) error {
	if resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return newApiErrorResponse(resp.StatusCode, resp.Status, body)
	}

	return nil
}
