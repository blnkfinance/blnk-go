package blnkgo_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupSearchService() (*MockClient, *blnkgo.SearchService) {
	mockClient := &MockClient{}
	svc := blnkgo.NewSearchService(mockClient)
	return mockClient, svc
}

func TestSearchService_SearchDocument_Success(t *testing.T) {
	mockClient, svc := setupSearchService()

	body := blnkgo.SearchParams{
		Q:       "test query",
		QueryBy: "field",
		Page:    1,
		PerPage: 10,
	}

	expectedResponse := &blnkgo.SearchResponse{
		Found: 1,
		OutOf: 1,
		Page:  1,
		RequestParams: blnkgo.SearchParams{
			Q:       "test query",
			QueryBy: "field",
			Page:    1,
			PerPage: 10,
		},
		SearchTimeMs: 100,
		Hits: []blnkgo.SearchHit{
			{
				Document: blnkgo.SearchDocument{
					BalanceID:     "balance123",
					Balance:       "100.0",
					CreditBalance: "50.0",
					DebitBalance:  "50.0",
					Currency:      "USD",
					Precision:     2,
					LedgerID:      "ledger123",
					CreatedAt:     blnkgo.FlexibleTime{Time: time.Now()},
					MetaData:      map[string]interface{}{"key": "value"},
				},
			},
		},
	}

	mockClient.On("NewRequest", "search/resource", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		searchResponse := args.Get(1).(*blnkgo.SearchResponse)
		*searchResponse = *expectedResponse
	})

	searchResponse, resp, err := svc.SearchDocument(body, "resource")

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, searchResponse)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestSearchService_SearchDocument_EmptyRequest(t *testing.T) {
	mockClient, svc := setupSearchService()
	body := blnkgo.SearchParams{}

	mockClient.On("NewRequest", "search/ledgers", http.MethodPost, body).Return(nil, fmt.Errorf("invalid request"))
	searchResponse, resp, err := svc.SearchDocument(body, "ledgers")

	assert.Error(t, err)
	assert.Nil(t, searchResponse)
	assert.Nil(t, resp)
	mockClient.AssertExpectations(t)
}

func TestSearchService_SearchDocument_ServerError(t *testing.T) {
	mockClient, svc := setupSearchService()
	body := blnkgo.SearchParams{
		Q:       "test query",
		QueryBy: "field",
		Page:    1,
		PerPage: 10,
	}

	expectedResp := &http.Response{StatusCode: http.StatusInternalServerError}

	mockClient.On("NewRequest", "search/ledgers", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(expectedResp, fmt.Errorf("server error"))

	searchResponse, resp, err := svc.SearchDocument(body, "ledgers")

	assert.Error(t, err)
	assert.Nil(t, searchResponse)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestSearchService_SearchDocument_InvalidResource(t *testing.T) {
	mockClient, svc := setupSearchService()
	body := blnkgo.SearchParams{
		Q:       "test query",
		QueryBy: "field",
		Page:    1,
		PerPage: 10,
	}

	mockClient.On("NewRequest", "search/invalid_resource", http.MethodPost, body).Return(nil, fmt.Errorf("invalid resource"))
	searchResponse, resp, err := svc.SearchDocument(body, "invalid_resource")

	assert.Error(t, err)
	assert.Nil(t, searchResponse)
	assert.Nil(t, resp)
	mockClient.AssertExpectations(t)
}

func TestSearchService_SearchDocument_EmptyResponse(t *testing.T) {
	mockClient, svc := setupSearchService()
	body := blnkgo.SearchParams{
		Q:       "test query",
		QueryBy: "field",
		Page:    1,
		PerPage: 10,
	}

	mockClient.On("NewRequest", "search/ledgers", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		searchResponse := args.Get(1).(*blnkgo.SearchResponse)
		*searchResponse = blnkgo.SearchResponse{}
	})

	searchResponse, resp, err := svc.SearchDocument(body, "ledgers")

	assert.NoError(t, err)
	assert.NotNil(t, searchResponse)
	assert.Equal(t, 0, searchResponse.Found)
	assert.Equal(t, 0, searchResponse.OutOf)
	assert.Equal(t, 0, searchResponse.Page)
	assert.Equal(t, 0, searchResponse.SearchTimeMs)
	assert.Empty(t, searchResponse.Hits)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestFlexibleTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name         string
		jsonData     string
		wantErr      bool
		expectedTime time.Time
	}{
		{
			name:         "Unix timestamp as number",
			jsonData:     `{"created_at": 1672531200}`,
			wantErr:      false,
			expectedTime: time.Unix(1672531200, 0),
		},
		{
			name:         "Unix timestamp as string",
			jsonData:     `{"created_at": "1672531200"}`,
			wantErr:      false,
			expectedTime: time.Unix(1672531200, 0),
		},
		{
			name:         "RFC3339 string",
			jsonData:     `{"created_at": "2023-01-01T00:00:00Z"}`,
			wantErr:      false,
			expectedTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:         "RFC3339 string with timezone",
			jsonData:     `{"created_at": "2023-01-01T12:30:45-03:00"}`,
			wantErr:      false,
			expectedTime: time.Date(2023, 1, 1, 15, 30, 45, 0, time.UTC),
		},
		{
			name:     "Invalid format",
			jsonData: `{"created_at": "invalid-date"}`,
			wantErr:  true,
		},
		{
			name:     "Empty string",
			jsonData: `{"created_at": ""}`,
			wantErr:  true,
		},
		{
			name:     "Null value",
			jsonData: `{"created_at": null}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var doc struct {
				CreatedAt blnkgo.FlexibleTime `json:"created_at"`
			}

			err := json.Unmarshal([]byte(tt.jsonData), &doc)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.False(t, doc.CreatedAt.Time.IsZero())

				// Verificar se o tempo foi parseado corretamente
				if !tt.expectedTime.IsZero() {
					assert.Equal(t, tt.expectedTime.Unix(), doc.CreatedAt.Time.Unix())
				}
			}
		})
	}
}

func TestFlexibleTime_MarshalJSON(t *testing.T) {
	testTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	flexTime := blnkgo.FlexibleTime{Time: testTime}

	data, err := json.Marshal(flexTime)
	assert.NoError(t, err)

	// Should marshal as Unix timestamp
	expected := fmt.Sprintf("%d", testTime.Unix())
	assert.Equal(t, expected, string(data))
}

func TestFlexibleTime_RoundTrip(t *testing.T) {
	originalTime := time.Date(2023, 8, 6, 15, 30, 45, 0, time.UTC)
	flexTime := blnkgo.FlexibleTime{Time: originalTime}

	// Marshal to JSON
	data, err := json.Marshal(flexTime)
	assert.NoError(t, err)

	// Unmarshal back
	var unmarshaled blnkgo.FlexibleTime
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	// Should be equal (considering Unix timestamp precision)
	assert.Equal(t, originalTime.Unix(), unmarshaled.Time.Unix())
}

func TestSearchDocument_MetaData_FlexibleTypes(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		wantErr  bool
	}{
		{
			name: "MetaData as object",
			jsonData: `{
				"balance_id": "bal-123",
				"balance": "100.50",
				"meta_data": {"key": "value", "count": 42}
			}`,
			wantErr: false,
		},
		{
			name: "MetaData as string",
			jsonData: `{
				"balance_id": "bal-123",
				"balance": "100.50",
				"meta_data": "string metadata"
			}`,
			wantErr: false,
		},
		{
			name: "MetaData as null",
			jsonData: `{
				"balance_id": "bal-123",
				"balance": "100.50",
				"meta_data": null
			}`,
			wantErr: false,
		},
		{
			name: "MetaData as array",
			jsonData: `{
				"balance_id": "bal-123",
				"balance": "100.50",
				"meta_data": ["item1", "item2"]
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var doc blnkgo.SearchDocument
			err := json.Unmarshal([]byte(tt.jsonData), &doc)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "bal-123", doc.BalanceID)
				assert.Equal(t, "100.50", doc.Balance)
				// MetaData pode ser qualquer tipo, então apenas verificamos que não é nil se não for explicitamente null
				if tt.name != "MetaData as null" {
					assert.NotNil(t, doc.MetaData)
				}
			}
		})
	}
}

func TestSearchDocument_TransactionFields(t *testing.T) {
	// JSON response similar to what's returned by the API for transactions
	transactionJSON := `{
		"allow_overdraft": true,
		"amount": 566,
		"amount_string": "566",
		"atomic": false,
		"created_at": 1754599843,
		"currency": "POINTS",
		"description": "Pontos transferidos do posto para o motorista",
		"destination": "bln_113a75b0-e838-48b6-934b-18b142295bb3",
		"destinations": [],
		"effective_date": 1754599900,
		"hash": "a872fd9adfe0173810b4d171360b98edc663bd9104a2db4dc53022e2deb348d2",
		"id": "26",
		"inflight": false,
		"inflight_expiry_date": 1754599843,
		"meta_data": "{\"QUEUED_PARENT_TRANSACTION\":\"txn_b1a740cc-5b8a-4370-b7f1-d4e4554a3029\",\"transaction_type\":\"posto -> motorista\"}",
		"overdraft_limit": 0,
		"parent_transaction": "txn_b1a740cc-5b8a-4370-b7f1-d4e4554a3029",
		"precise_amount": "566",
		"precision": 1,
		"rate": 1,
		"reference": "motor-test-34c9737f-1bc8-4495-a33d-0d8207be46c3_q",
		"scheduled_for": 1754599843,
		"skip_queue": false,
		"source": "bln_f7e6fbc5-ddac-4b79-adf0-151cc7f9605e",
		"sources": [],
		"status": "APPLIED",
		"transaction_id": "txn_2dd81e34-c72b-4467-8dbe-e3f126a73e92"
	}`

	var doc blnkgo.SearchDocument
	err := json.Unmarshal([]byte(transactionJSON), &doc)

	assert.NoError(t, err)

	// Test transaction-specific fields
	assert.Equal(t, "txn_2dd81e34-c72b-4467-8dbe-e3f126a73e92", doc.TransactionID)
	assert.Equal(t, 566.0, doc.Amount)
	assert.Equal(t, "566", doc.AmountString)
	assert.Equal(t, "bln_f7e6fbc5-ddac-4b79-adf0-151cc7f9605e", doc.Source)
	assert.Equal(t, "bln_113a75b0-e838-48b6-934b-18b142295bb3", doc.Destination)
	assert.Equal(t, "APPLIED", doc.Status)
	assert.Equal(t, "txn_b1a740cc-5b8a-4370-b7f1-d4e4554a3029", doc.ParentTransaction)
	assert.Equal(t, "a872fd9adfe0173810b4d171360b98edc663bd9104a2db4dc53022e2deb348d2", doc.Hash)
	assert.Equal(t, false, doc.Atomic)
	assert.Equal(t, false, doc.Inflight)
	assert.Equal(t, true, doc.AllowOverdraft)
	assert.Equal(t, 0.0, doc.OverdraftLimit)
	assert.Equal(t, "566", doc.PreciseAmount)
	assert.Equal(t, 1, doc.Precision)
	assert.Equal(t, 1.0, doc.Rate)
	assert.Equal(t, "motor-test-34c9737f-1bc8-4495-a33d-0d8207be46c3_q", doc.Reference)
	assert.Equal(t, false, doc.SkipQueue)
	assert.Equal(t, "26", doc.ID)
	assert.Equal(t, "POINTS", doc.Currency)
	assert.Equal(t, "Pontos transferidos do posto para o motorista", doc.Description)

	// Test common fields
	assert.NotNil(t, doc.MetaData)
	assert.NotZero(t, doc.CreatedAt.Time)

	// Test that time fields were parsed correctly
	expectedTime := time.Unix(1754599843, 0)
	expectedEffectiveDate := time.Unix(1754599900, 0)
	assert.Equal(t, expectedTime, doc.CreatedAt.Time)
	assert.Equal(t, expectedTime, doc.ScheduledFor.Time)
	assert.Equal(t, expectedTime, doc.InflightExpiryDate.Time)
	assert.Equal(t, expectedEffectiveDate, doc.EffectiveDate.Time)
}

func TestSearchDocument_EffectiveDate_Parsing(t *testing.T) {
	tests := []struct {
		name         string
		jsonData     string
		wantErr      bool
		expectedTime time.Time
	}{
		{
			name: "EffectiveDate as Unix timestamp",
			jsonData: `{
				"transaction_id": "txn_123",
				"effective_date": 1754599900
			}`,
			wantErr:      false,
			expectedTime: time.Unix(1754599900, 0),
		},
		{
			name: "EffectiveDate as RFC3339 string",
			jsonData: `{
				"transaction_id": "txn_123",
				"effective_date": "2023-08-15T10:30:00Z"
			}`,
			wantErr:      false,
			expectedTime: time.Date(2023, 8, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name: "EffectiveDate as Unix timestamp string",
			jsonData: `{
				"transaction_id": "txn_123",
				"effective_date": "1754599900"
			}`,
			wantErr:      false,
			expectedTime: time.Unix(1754599900, 0),
		},
		{
			name: "EffectiveDate omitted",
			jsonData: `{
				"transaction_id": "txn_123",
				"amount": 100.0
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var doc blnkgo.SearchDocument
			err := json.Unmarshal([]byte(tt.jsonData), &doc)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "txn_123", doc.TransactionID)

				if !tt.expectedTime.IsZero() {
					assert.Equal(t, tt.expectedTime.Unix(), doc.EffectiveDate.Time.Unix())
				} else {
					// If no effective_date is provided, it should be zero time
					assert.True(t, doc.EffectiveDate.Time.IsZero())
				}
			}
		})
	}
}

func TestSearchDocument_LedgerFields(t *testing.T) {
	// JSON response similar to what's returned by the API for ledgers
	ledgerJSON := `{
		"created_at": 1754599640,
		"id": "ldg_40688495-864f-4442-ac37-68dba582b755",
		"ledger_id": "ldg_40688495-864f-4442-ac37-68dba582b755",
		"meta_data": "{\"description\":\"motorista pontos\"}",
		"name": "Motorista (destino)"
	}`

	var doc blnkgo.SearchDocument
	err := json.Unmarshal([]byte(ledgerJSON), &doc)

	assert.NoError(t, err)

	// Test ledger-specific fields
	assert.Equal(t, "ldg_40688495-864f-4442-ac37-68dba582b755", doc.ID)
	assert.Equal(t, "ldg_40688495-864f-4442-ac37-68dba582b755", doc.LedgerID)
	assert.Equal(t, "Motorista (destino)", doc.Name)

	// Test common fields
	assert.NotNil(t, doc.MetaData)
	assert.NotZero(t, doc.CreatedAt.Time)

	// Test that time field was parsed correctly
	expectedTime := time.Unix(1754599640, 0)
	assert.Equal(t, expectedTime, doc.CreatedAt.Time)

	// Test that metadata is correctly parsed as string
	if metaStr, ok := doc.MetaData.(string); ok {
		assert.Contains(t, metaStr, "motorista pontos")
	}
}
