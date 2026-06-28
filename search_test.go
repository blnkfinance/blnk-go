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

func intPtr(v int) *int {
	return &v
}

func TestSearchService_SearchDocument_Success(t *testing.T) {
	mockClient, svc := setupSearchService()

	body := blnkgo.SearchParams{
		Q:          "test query",
		QueryBy:    "field",
		Page:       1,
		PerPage:    10,
		GroupBy:    "",
		GroupLimit: 1,
	}

	expectedResponse := &blnkgo.SearchResponse{
		Found: 1,
		OutOf: 1,
		Page:  1,
		RequestParams: blnkgo.SearchParams{
			Q:          "test query",
			QueryBy:    "field",
			Page:       1,
			PerPage:    10,
			GroupBy:    "",
			GroupLimit: 1,
		},
		SearchTimeMs: 100,
		Hits: []blnkgo.SearchHit{
			{
				Document: blnkgo.SearchDocument{
					BalanceID:             "balance123",
					Balance:               "100.0",
					CreditBalance:         "50.0",
					DebitBalance:          "50.0",
					Currency:              "USD",
					Precision:             2,
					LedgerID:              "ledger123",
					InflightBalance:       "25.0",
					InflightCreditBalance: "15.0",
					InflightDebitBalance:  "10.0",
					CreatedAt:             blnkgo.FlexibleTime{Time: time.Now()},
					MetaData:              map[string]interface{}{"key": "value"},
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

func TestSearchService_SearchDocument_IdentitiesResource(t *testing.T) {
	mockClient, svc := setupSearchService()

	body := blnkgo.SearchParams{
		Q:       "john.doe@example.com",
		QueryBy: "email_address",
		Page:    1,
		PerPage: 10,
	}

	mockClient.On("NewRequest", "search/identities", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		searchResponse := args.Get(1).(*blnkgo.SearchResponse)
		*searchResponse = blnkgo.SearchResponse{Found: 1, Page: 1}
	})

	searchResponse, resp, err := svc.SearchDocument(body, blnkgo.Identities)

	assert.NoError(t, err)
	assert.NotNil(t, searchResponse)
	assert.Equal(t, 1, searchResponse.Found)
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
		Q:          "test query",
		QueryBy:    "field",
		Page:       1,
		PerPage:    10,
		GroupBy:    "",
		GroupLimit: 1,
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
		Q:          "test query",
		QueryBy:    "field",
		Page:       1,
		PerPage:    10,
		GroupBy:    "",
		GroupLimit: 1,
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
		Q:          "test query",
		QueryBy:    "field",
		Page:       1,
		PerPage:    10,
		GroupBy:    "",
		GroupLimit: 1,
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
	assert.Empty(t, searchResponse.GroupedHits)
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

func TestSearchDocument_BalanceInflightFields(t *testing.T) {
	balanceJSON := `{
		"balance_id": "bal-123",
		"balance": "100.50",
		"credit_balance": "75.25",
		"debit_balance": "25.25",
		"inflight_balance": "10.00",
		"inflight_credit_balance": "7.50",
		"inflight_debit_balance": "2.50",
		"currency": "USD",
		"ledger_id": "ledger-123",
		"created_at": 1672531200,
		"meta_data": {"owner": "customer-1"}
	}`

	var doc blnkgo.SearchDocument
	err := json.Unmarshal([]byte(balanceJSON), &doc)

	assert.NoError(t, err)
	assert.Equal(t, "bal-123", doc.BalanceID)
	assert.Equal(t, "100.50", doc.Balance)
	assert.Equal(t, "75.25", doc.CreditBalance)
	assert.Equal(t, "25.25", doc.DebitBalance)
	assert.Equal(t, "10.00", doc.InflightBalance)
	assert.Equal(t, "7.50", doc.InflightCreditBalance)
	assert.Equal(t, "2.50", doc.InflightDebitBalance)
	assert.Equal(t, "USD", doc.Currency)
	assert.Equal(t, "ledger-123", doc.LedgerID)
	assert.Equal(t, time.Unix(1672531200, 0), doc.CreatedAt.Time)
	assert.NotNil(t, doc.MetaData)
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
					assert.True(t, doc.EffectiveDate.Time.IsZero())
				}
			}
		})
	}
}

func TestSearchService_SearchDocument_WithGroupedHits(t *testing.T) {
	mockClient, svc := setupSearchService()

	body := blnkgo.SearchParams{
		Q:          "bln_ad4980a0-415a-4fd4-a800-3eabbb48b090",
		QueryBy:    "destination",
		GroupBy:    "parent_transaction",
		GroupLimit: 1,
		Page:       1,
		PerPage:    10,
	}

	expectedResponse := &blnkgo.SearchResponse{
		Found: 29,
		OutOf: 1948,
		Page:  1,
		RequestParams: blnkgo.SearchParams{
			Q:       "bln_ad4980a0-415a-4fd4-a800-3eabbb48b090",
			PerPage: 10,
		},
		SearchTimeMs: 1,
		GroupedHits: []blnkgo.GroupedHit{
			{
				GroupKey: []string{"txn_7c779669-da6a-4bfb-ac03-a53dc7c5c568"},
				Hits: []blnkgo.SearchHit{
					{
						Document: blnkgo.SearchDocument{
							ID:                "txn_14706192-e53d-43cc-86a1-b8807a09b351",
							TransactionID:     "txn_14706192-e53d-43cc-86a1-b8807a09b351",
							Amount:            12.31,
							AmountString:      "12.31",
							Currency:          "BRL",
							Description:       "Um uid 4063248968213714",
							Destination:       "bln_ad4980a0-415a-4fd4-a800-3eabbb48b090",
							Source:            "bln_d17a7c1f-6f0e-47b9-9360-aeee6f0f247e",
							Status:            "APPLIED",
							ParentTransaction: "txn_7c779669-da6a-4bfb-ac03-a53dc7c5c568",
							CreatedAt:         blnkgo.FlexibleTime{Time: time.Unix(1764873061, 0)},
						},
					},
				},
			},
			{
				GroupKey: []string{"txn_56bc1ffd-d7d9-443d-8011-7c4fbc7e6677"},
				Hits: []blnkgo.SearchHit{
					{
						Document: blnkgo.SearchDocument{
							ID:                "txn_4b0a842c-2066-4097-8b4b-ed481fb269b1",
							TransactionID:     "txn_4b0a842c-2066-4097-8b4b-ed481fb269b1",
							Amount:            14.12,
							AmountString:      "14.12",
							Currency:          "BRL",
							Destination:       "bln_ad4980a0-415a-4fd4-a800-3eabbb48b090",
							Source:            "bln_d17a7c1f-6f0e-47b9-9360-aeee6f0f247e",
							Status:            "APPLIED",
							ParentTransaction: "txn_56bc1ffd-d7d9-443d-8011-7c4fbc7e6677",
							CreatedAt:         blnkgo.FlexibleTime{Time: time.Unix(1764872164, 0)},
						},
					},
				},
			},
		},
	}

	mockClient.On("NewRequest", "search/transactions", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		searchResponse := args.Get(1).(*blnkgo.SearchResponse)
		*searchResponse = *expectedResponse
	})

	searchResponse, resp, err := svc.SearchDocument(body, "transactions")

	assert.NoError(t, err)
	assert.NotNil(t, searchResponse)
	assert.Equal(t, 29, searchResponse.Found)
	assert.Equal(t, 1948, searchResponse.OutOf)
	assert.Equal(t, 1, searchResponse.Page)
	assert.Equal(t, 1, searchResponse.SearchTimeMs)
	assert.Empty(t, searchResponse.Hits)
	assert.NotEmpty(t, searchResponse.GroupedHits)
	assert.Len(t, searchResponse.GroupedHits, 2)

	// Verify first group
	firstGroup := searchResponse.GroupedHits[0]
	assert.Equal(t, []string{"txn_7c779669-da6a-4bfb-ac03-a53dc7c5c568"}, firstGroup.GroupKey)
	assert.Len(t, firstGroup.Hits, 1)
	assert.Equal(t, "txn_14706192-e53d-43cc-86a1-b8807a09b351", firstGroup.Hits[0].Document.TransactionID)
	assert.Equal(t, "txn_7c779669-da6a-4bfb-ac03-a53dc7c5c568", firstGroup.Hits[0].Document.ParentTransaction)

	// Verify second group
	secondGroup := searchResponse.GroupedHits[1]
	assert.Equal(t, []string{"txn_56bc1ffd-d7d9-443d-8011-7c4fbc7e6677"}, secondGroup.GroupKey)
	assert.Len(t, secondGroup.Hits, 1)
	assert.Equal(t, "txn_4b0a842c-2066-4097-8b4b-ed481fb269b1", secondGroup.Hits[0].Document.TransactionID)
	assert.Equal(t, "txn_56bc1ffd-d7d9-443d-8011-7c4fbc7e6677", secondGroup.Hits[0].Document.ParentTransaction)

	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestSearchService_SearchDocument_WithRegularHits(t *testing.T) {
	mockClient, svc := setupSearchService()

	body := blnkgo.SearchParams{
		Q:       "bln_ad4980a0-415a-4fd4-a800-3eabbb48b090",
		QueryBy: "destination",
		Page:    1,
		PerPage: 10,
	}

	expectedResponse := &blnkgo.SearchResponse{
		Found: 61,
		OutOf: 0,
		Page:  1,
		RequestParams: blnkgo.SearchParams{
			Q:       "bln_ad4980a0-415a-4fd4-a800-3eabbb48b090",
			PerPage: 10,
		},
		SearchTimeMs: 1,
		Hits: []blnkgo.SearchHit{
			{
				Document: blnkgo.SearchDocument{
					ID:                "txn_14706192-e53d-43cc-86a1-b8807a09b351",
					TransactionID:     "txn_14706192-e53d-43cc-86a1-b8807a09b351",
					Amount:            12.31,
					AmountString:      "12.31",
					Currency:          "BRL",
					Description:       "Um uid 4063248968213714",
					Destination:       "bln_ad4980a0-415a-4fd4-a800-3eabbb48b090",
					Source:            "bln_d17a7c1f-6f0e-47b9-9360-aeee6f0f247e",
					Status:            "APPLIED",
					ParentTransaction: "txn_7c779669-da6a-4bfb-ac03-a53dc7c5c568",
					CreatedAt:         blnkgo.FlexibleTime{Time: time.Unix(1764873061, 0)},
				},
			},
			{
				Document: blnkgo.SearchDocument{
					ID:                "txn_8ad25be5-e25d-4f11-bc79-d1f208f0a7e0",
					TransactionID:     "txn_8ad25be5-e25d-4f11-bc79-d1f208f0a7e0",
					Amount:            12.31,
					AmountString:      "12.31",
					Currency:          "BRL",
					Destination:       "bln_ad4980a0-415a-4fd4-a800-3eabbb48b090",
					Source:            "bln_d17a7c1f-6f0e-47b9-9360-aeee6f0f247e",
					Status:            "INFLIGHT",
					ParentTransaction: "txn_7c779669-da6a-4bfb-ac03-a53dc7c5c568",
					Inflight:          true,
					CreatedAt:         blnkgo.FlexibleTime{Time: time.Unix(1764873018, 0)},
				},
			},
		},
	}

	mockClient.On("NewRequest", "search/transactions", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		searchResponse := args.Get(1).(*blnkgo.SearchResponse)
		*searchResponse = *expectedResponse
	})

	searchResponse, resp, err := svc.SearchDocument(body, "transactions")

	assert.NoError(t, err)
	assert.NotNil(t, searchResponse)
	assert.Equal(t, 61, searchResponse.Found)
	assert.Equal(t, 0, searchResponse.OutOf)
	assert.Equal(t, 1, searchResponse.Page)
	assert.Equal(t, 1, searchResponse.SearchTimeMs)
	assert.NotEmpty(t, searchResponse.Hits)
	assert.Empty(t, searchResponse.GroupedHits)
	assert.Len(t, searchResponse.Hits, 2)

	// Verify first hit
	firstHit := searchResponse.Hits[0]
	assert.Equal(t, "txn_14706192-e53d-43cc-86a1-b8807a09b351", firstHit.Document.TransactionID)
	assert.Equal(t, "APPLIED", firstHit.Document.Status)
	assert.Equal(t, 12.31, firstHit.Document.Amount)

	// Verify second hit
	secondHit := searchResponse.Hits[1]
	assert.Equal(t, "txn_8ad25be5-e25d-4f11-bc79-d1f208f0a7e0", secondHit.Document.TransactionID)
	assert.Equal(t, "INFLIGHT", secondHit.Document.Status)
	assert.True(t, secondHit.Document.Inflight)

	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestSearchDocument_LedgerFields(t *testing.T) {
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

	assert.Equal(t, "ldg_40688495-864f-4442-ac37-68dba582b755", doc.ID)
	assert.Equal(t, "ldg_40688495-864f-4442-ac37-68dba582b755", doc.LedgerID)
	assert.Equal(t, "Motorista (destino)", doc.Name)

	assert.NotNil(t, doc.MetaData)
	assert.NotZero(t, doc.CreatedAt.Time)

	expectedTime := time.Unix(1754599640, 0)
	assert.Equal(t, expectedTime, doc.CreatedAt.Time)

	if metaStr, ok := doc.MetaData.(string); ok {
		assert.Contains(t, metaStr, "motorista pontos")
	}
}

func TestSearchService_StartReindex_SuccessEmptyBody(t *testing.T) {
	mockClient, svc := setupSearchService()

	mockClient.On("NewRequest", "search/reindex", http.MethodPost, struct{}{}).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusAccepted}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.StartReindexResponse)
		*resp = blnkgo.StartReindexResponse{
			Message: "Reindex operation started",
			Progress: blnkgo.ReindexProgress{
				Status: "in_progress",
				Phase:  "indexing_transactions",
			},
		}
	})

	started, httpResp, err := svc.StartReindex(nil)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusAccepted, httpResp.StatusCode)
	assert.Equal(t, "Reindex operation started", started.Message)
	assert.Equal(t, "in_progress", started.Progress.Status)
	mockClient.AssertExpectations(t)
}

func TestSearchService_StartReindex_SuccessWithBatchSize(t *testing.T) {
	mockClient, svc := setupSearchService()

	opts := &blnkgo.StartReindexRequest{BatchSize: intPtr(500)}
	body := blnkgo.StartReindexRequest{BatchSize: intPtr(500)}

	mockClient.On("NewRequest", "search/reindex", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusAccepted}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.StartReindexResponse)
		*resp = blnkgo.StartReindexResponse{Message: "Reindex operation started"}
	})

	started, httpResp, err := svc.StartReindex(opts)

	assert.NoError(t, err)
	assert.NotNil(t, started)
	assert.Equal(t, http.StatusAccepted, httpResp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestSearchService_StartReindex_ValidationError(t *testing.T) {
	mockClient, svc := setupSearchService()

	opts := &blnkgo.StartReindexRequest{BatchSize: intPtr(0)}
	_, _, err := svc.StartReindex(opts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "batch_size must be a positive integer")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestSearchService_StartReindex_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupSearchService()

	mockClient.On("NewRequest", "search/reindex", http.MethodPost, struct{}{}).Return(nil, fmt.Errorf("failed to create request"))

	started, httpResp, err := svc.StartReindex(nil)

	assert.Error(t, err)
	assert.Nil(t, started)
	assert.Nil(t, httpResp)
	mockClient.AssertExpectations(t)
}

func TestSearchService_GetReindexStatus_Success(t *testing.T) {
	mockClient, svc := setupSearchService()

	mockClient.On("NewRequest", "search/reindex", http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.ReindexProgress)
		*resp = blnkgo.ReindexProgress{
			Status: "in_progress",
			Phase:  "indexing_transactions",
		}
	})

	progress, httpResp, err := svc.GetReindexStatus()

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, "in_progress", progress.Status)
	assert.Equal(t, "indexing_transactions", progress.Phase)
	mockClient.AssertExpectations(t)
}

func TestSearchService_GetReindexStatus_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupSearchService()

	mockClient.On("NewRequest", "search/reindex", http.MethodGet, nil).Return(nil, fmt.Errorf("failed to create request"))

	progress, httpResp, err := svc.GetReindexStatus()

	assert.Error(t, err)
	assert.Nil(t, progress)
	assert.Nil(t, httpResp)
	mockClient.AssertExpectations(t)
}
