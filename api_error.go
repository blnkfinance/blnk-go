package blnkgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Stable Core error_detail.code values called out for SDK 1.4.0 / Core 0.15.3.
// Compare ErrorDetail.Code — do not branch on message text.
//
// https://docs.blnkfinance.com/advanced/error-codes
const (
	// ErrorCodeTxnInvalidAmount is the catalog code for invalid amounts.
	ErrorCodeTxnInvalidAmount = "TXN_INVALID_AMOUNT"
	// ErrorCodeGenConflict is returned for a duplicate internal-balance
	// indicator + currency (HTTP 409).
	ErrorCodeGenConflict = "GEN_CONFLICT"
	// ErrorCodeTxnValidationError is returned when Core rejects a request
	// before processing. On 0.15.3 this includes negative amount/precision
	// and source equal to destination.
	ErrorCodeTxnValidationError = "TXN_VALIDATION_ERROR"
)

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
