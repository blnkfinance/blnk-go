package blnkgo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type SearchService service

// FlexibleTime is a custom time type that can unmarshal from both Unix timestamp (int64) and RFC3339 string
type FlexibleTime struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler interface
func (ft *FlexibleTime) UnmarshalJSON(data []byte) error {
	// Handle null explicitly
	if string(data) == "null" {
		return fmt.Errorf("cannot parse time from null value")
	}

	// Try to unmarshal as Unix timestamp (number)
	var timestamp int64
	if err := json.Unmarshal(data, &timestamp); err == nil {
		ft.Time = time.Unix(timestamp, 0)
		return nil
	}

	// Try to unmarshal as string
	var timeStr string
	if err := json.Unmarshal(data, &timeStr); err == nil {
		// Handle empty string
		if timeStr == "" {
			return fmt.Errorf("cannot parse time from empty string")
		}

		// Try parsing as RFC3339
		if parsedTime, err := time.Parse(time.RFC3339, timeStr); err == nil {
			ft.Time = parsedTime
			return nil
		}

		// Try parsing as Unix timestamp string
		if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
			ft.Time = time.Unix(timestamp, 0)
			return nil
		}
	}

	return fmt.Errorf("cannot parse time from: %s", string(data))
}

// MarshalJSON implements json.Marshaler interface
func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ft.Time.Unix())
}

type SearchParams struct {
	Q        string `json:"q"`
	QueryBy  string `json:"query_by,omitempty"`
	FilterBy string `json:"filter_by,omitempty"`
	SortBy   string `json:"sort_by,omitempty"`
	Page     int    `json:"page,omitempty"`
	PerPage  int    `json:"per_page,omitempty"`
}

type SearchResponse struct {
	Found         int          `json:"found"`
	OutOf         int          `json:"out_of"`
	Page          int          `json:"page"`
	RequestParams SearchParams `json:"request_params"`
	SearchTimeMs  int          `json:"search_time_ms"`
	Hits          []SearchHit  `json:"hits"`
}

type SearchHit struct {
	Document SearchDocument `json:"document"`
}

type SearchDocument struct {
	// Common fields
	ID        string       `json:"id,omitempty"`
	CreatedAt FlexibleTime `json:"created_at"` // Accepts both Unix timestamp and RFC3339 string
	MetaData  interface{}  `json:"meta_data"`  // Can be string, map, or other types

	// Balance fields
	BalanceID     string `json:"balance_id,omitempty"`
	Balance       string `json:"balance,omitempty"`
	CreditBalance string `json:"credit_balance,omitempty"`
	DebitBalance  string `json:"debit_balance,omitempty"`
	Currency      string `json:"currency,omitempty"`
	Precision     int    `json:"precision,omitempty"`
	LedgerID      string `json:"ledger_id,omitempty"`

	// Transaction fields
	TransactionID      string       `json:"transaction_id,omitempty"`
	Amount             float64      `json:"amount,omitempty"`
	AmountString       string       `json:"amount_string,omitempty"`
	Source             string       `json:"source,omitempty"`
	Destination        string       `json:"destination,omitempty"`
	Sources            interface{}  `json:"sources,omitempty"`
	Destinations       interface{}  `json:"destinations,omitempty"`
	Reference          string       `json:"reference,omitempty"`
	Description        string       `json:"description,omitempty"`
	Status             string       `json:"status,omitempty"`
	ParentTransaction  string       `json:"parent_transaction,omitempty"`
	Hash               string       `json:"hash,omitempty"`
	Atomic             bool         `json:"atomic,omitempty"`
	Inflight           bool         `json:"inflight,omitempty"`
	AllowOverdraft     bool         `json:"allow_overdraft,omitempty"`
	OverdraftLimit     float64      `json:"overdraft_limit,omitempty"`
	ScheduledFor       FlexibleTime `json:"scheduled_for,omitempty"`
	InflightExpiryDate FlexibleTime `json:"inflight_expiry_date,omitempty"`
	SkipQueue          bool         `json:"skip_queue,omitempty"`
	Rate               float64      `json:"rate,omitempty"`
	PreciseAmount      string       `json:"precise_amount,omitempty"`
	EffectiveDate      FlexibleTime `json:"effective_date,omitempty"`

	// Ledger fields
	Name string `json:"name,omitempty"`
}

func (s *SearchService) SearchDocument(body SearchParams, resource ResourceType) (*SearchResponse, *http.Response, error) {
	u := fmt.Sprintf("search/%s", resource)
	req, err := s.client.NewRequest(u, http.MethodPost, body)
	if err != nil {
		return nil, nil, err
	}

	searchResponse := new(SearchResponse)
	resp, err := s.client.CallWithRetry(req, searchResponse)
	if err != nil {
		return nil, resp, err
	}

	return searchResponse, resp, nil
}

func NewSearchService(c ClientInterface) *SearchService {
	return &SearchService{client: c}
}
