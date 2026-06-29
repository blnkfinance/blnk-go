package blnkgo_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseApiErrorBody_StructuredErrorDetail(t *testing.T) {
	body := []byte(`{
		"error": "transaction not found",
		"error_detail": {
			"code": "TXN_NOT_FOUND",
			"message": "transaction not found",
			"details": {}
		}
	}`)

	legacy, detail := blnkgo.ParseApiErrorBody(body)
	require.NotNil(t, detail)
	assert.Equal(t, "transaction not found", legacy)
	assert.Equal(t, "TXN_NOT_FOUND", detail.Code)
	assert.Equal(t, "transaction not found", detail.Message)
	assert.NotNil(t, detail.Details)
}

func TestParseApiErrorBody_ConflictWithDetails(t *testing.T) {
	body := []byte(`{
		"error": "duplicate reference",
		"error_detail": {
			"code": "TXN_DUPLICATE_REFERENCE",
			"message": "duplicate reference",
			"details": {"reference": "ref_001"}
		}
	}`)

	_, detail := blnkgo.ParseApiErrorBody(body)
	require.NotNil(t, detail)
	assert.Equal(t, "TXN_DUPLICATE_REFERENCE", detail.Code)
	assert.Equal(t, "duplicate reference", detail.Message)
}

func TestParseApiErrorBody_LegacyErrorOnly(t *testing.T) {
	body := []byte(`{"error": "ledger not found"}`)

	legacy, detail := blnkgo.ParseApiErrorBody(body)
	assert.Equal(t, "ledger not found", legacy)
	require.NotNil(t, detail)
	assert.Equal(t, "UNKNOWN", detail.Code)
	assert.Equal(t, "ledger not found", detail.Message)
}

func TestParseApiErrorBody_EmptyBody(t *testing.T) {
	legacy, detail := blnkgo.ParseApiErrorBody(nil)
	assert.Empty(t, legacy)
	assert.Nil(t, detail)
}

func TestParseApiErrorBody_InvalidJSON(t *testing.T) {
	legacy, detail := blnkgo.ParseApiErrorBody([]byte(`not json`))
	assert.Empty(t, legacy)
	assert.Nil(t, detail)
}

func TestCheckResponse_AttachesErrorDetail(t *testing.T) {
	client := &blnkgo.Client{}
	resp := &http.Response{
		StatusCode: http.StatusNotFound,
		Status:     "404 Not Found",
		Body: io.NopCloser(strings.NewReader(`{
			"error": "transaction not found",
			"error_detail": {
				"code": "TXN_NOT_FOUND",
				"message": "transaction not found",
				"details": {}
			}
		}`)),
	}

	err := client.CheckResponse(resp)
	require.Error(t, err)

	apiErr, ok := blnkgo.AsApiErrorResponse(err)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.Status)
	assert.Equal(t, "transaction not found", apiErr.LegacyError)
	require.NotNil(t, apiErr.ErrorDetail)
	assert.Equal(t, "TXN_NOT_FOUND", apiErr.ErrorDetail.Code)
}

func TestCheckResponse_LockedErrorDetail(t *testing.T) {
	client := &blnkgo.Client{}
	resp := &http.Response{
		StatusCode: http.StatusLocked,
		Status:     "423 Locked",
		Body: io.NopCloser(strings.NewReader(`{
			"error": "resource locked",
			"error_detail": {
				"code": "GEN_RESOURCE_LOCKED",
				"message": "resource locked",
				"details": {}
			}
		}`)),
	}

	err := client.CheckResponse(resp)
	require.Error(t, err)

	apiErr, ok := blnkgo.AsApiErrorResponse(err)
	require.True(t, ok)
	require.NotNil(t, apiErr.ErrorDetail)
	assert.Equal(t, "GEN_RESOURCE_LOCKED", apiErr.ErrorDetail.Code)
}

func TestApiErrorResponse_ErrorStringUsesCode(t *testing.T) {
	apiErr := &blnkgo.ApiErrorResponse{
		Status: 409,
		ErrorDetail: &blnkgo.ApiErrorDetail{
			Code:    "GEN_CONFLICT",
			Message: "conflict",
		},
	}
	assert.Contains(t, apiErr.Error(), "GEN_CONFLICT")
	assert.Contains(t, apiErr.Error(), "conflict")
}
