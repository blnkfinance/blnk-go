package blnkgo

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"strconv"
)

type BalanceMonitorService service

// MonitorTrigger controls how often a balance monitor fires (Core 0.15.7+).
type MonitorTrigger string

const (
	// MonitorTriggerEdge fires once when the condition becomes true, and again
	// only after it has gone false and become true again. Core's default.
	MonitorTriggerEdge MonitorTrigger = "edge"
	// MonitorTriggerLevel fires on every balance update while the condition holds.
	MonitorTriggerLevel MonitorTrigger = "level"
)

// MonitorCondition is a balance monitor's rule. The threshold is in major
// units (Core stores Value × Precision as PreciseValue).
//
// Value holds a whole-number threshold. For a fractional one such as 100.50,
// set DecimalValue instead; when DecimalValue is set it is sent in place of
// Value. On reads, a fractional threshold fills DecimalValue and leaves the
// whole part in Value. Use Threshold to read either.
type MonitorCondition struct {
	Field     string                    `json:"field"`
	Operator  MonitorConditionOperators `json:"operator"`
	Value     int64                     `json:"value"`
	Precision int64                     `json:"precision"`
	// DecimalValue is a fractional threshold in major units (Core 0.15.x
	// accepts values such as 100.50). It takes precedence over Value.
	DecimalValue *float64 `json:"-"`
	// PreciseValue is the threshold in minor units (Value × Precision). Core
	// computes it and returns it on reads; it is ignored on requests.
	PreciseValue *big.Int `json:"precise_value,omitempty"`
}

// Threshold returns the threshold in major units: DecimalValue when set,
// otherwise Value.
func (c MonitorCondition) Threshold() float64 {
	if c.DecimalValue != nil {
		return *c.DecimalValue
	}
	return float64(c.Value)
}

type monitorConditionWire struct {
	Field        string                    `json:"field"`
	Operator     MonitorConditionOperators `json:"operator"`
	Value        json.Number               `json:"value"`
	Precision    int64                     `json:"precision"`
	PreciseValue *big.Int                  `json:"precise_value,omitempty"`
}

// MarshalJSON sends DecimalValue when set, otherwise Value. The keys sent are
// the same as before DecimalValue existed.
func (c MonitorCondition) MarshalJSON() ([]byte, error) {
	value := strconv.FormatInt(c.Value, 10)
	if c.DecimalValue != nil {
		d := *c.DecimalValue
		if math.IsNaN(d) || math.IsInf(d, 0) {
			return nil, fmt.Errorf("monitor condition value must be a finite number")
		}
		value = strconv.FormatFloat(d, 'f', -1, 64)
	}
	return json.Marshal(monitorConditionWire{
		Field:        c.Field,
		Operator:     c.Operator,
		Value:        json.Number(value),
		Precision:    c.Precision,
		PreciseValue: c.PreciseValue,
	})
}

// UnmarshalJSON accepts whole and fractional thresholds. A whole number fills
// Value only; a fractional one fills DecimalValue and the whole part of Value.
func (c *MonitorCondition) UnmarshalJSON(data []byte) error {
	var w monitorConditionWire
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	*c = MonitorCondition{
		Field:        w.Field,
		Operator:     w.Operator,
		Precision:    w.Precision,
		PreciseValue: w.PreciseValue,
	}
	if w.Value == "" {
		return nil
	}
	if n, err := w.Value.Int64(); err == nil {
		c.Value = n
		return nil
	}
	f, err := w.Value.Float64()
	if err != nil {
		return fmt.Errorf("monitor condition value %q: %w", w.Value, err)
	}
	if f == math.Trunc(f) && math.Abs(f) < 1<<63 {
		c.Value = int64(f)
		return nil
	}
	c.Value = int64(math.Trunc(f))
	c.DecimalValue = &f
	return nil
}

// MonitorData represents the data structure for monitoring information.
type MonitorData struct {
	Condition   MonitorCondition `json:"condition"`
	Description string           `json:"description,omitempty"`
	BalanceID   string           `json:"balance_id"`
	CallBackURL string           `json:"call_back_url,omitempty"`
	// Trigger is MonitorTriggerEdge or MonitorTriggerLevel. Leave it empty for
	// Core's default (edge).
	Trigger MonitorTrigger `json:"trigger,omitempty"`
}

// MonitorDataResp extends MonitorData with additional fields for response data.
type MonitorDataResp struct {
	MonitorData
	MonitorID string `json:"monitor_id"`
	CreatedAt string `json:"created_at"` // ISO date string
	// ConditionState is the last observed truth value of an edge-triggered
	// monitor's condition. Nil until Core has evaluated it.
	ConditionState *bool `json:"condition_state,omitempty"`
}

// UpdateMonitorRequest is the body for UpdatePartial. Core 0.15.9 keeps any
// field the request leaves out, so set only what should change. Nil pointers
// and an empty Trigger are omitted from the JSON.
//
// The callback URL cannot be changed after creation; Core ignores it on update.
type UpdateMonitorRequest struct {
	BalanceID *string `json:"balance_id,omitempty"`
	// Description set to a pointer to "" clears the description.
	Description *string                 `json:"description,omitempty"`
	Trigger     MonitorTrigger          `json:"trigger,omitempty"`
	Condition   *UpdateMonitorCondition `json:"condition,omitempty"`
}

// UpdateMonitorCondition holds the condition fields to change. A Value that
// points at 0 moves the threshold to zero; a nil Value keeps the stored one.
// Core ignores a Precision change unless Value is sent with it.
type UpdateMonitorCondition struct {
	Field     *string                    `json:"field,omitempty"`
	Operator  *MonitorConditionOperators `json:"operator,omitempty"`
	Value     *float64                   `json:"value,omitempty"`
	Precision *int64                     `json:"precision,omitempty"`
}

// UpdateBalanceMonitorResponse is Core's reply to PUT /balance-monitors/{id}.
type UpdateBalanceMonitorResponse struct {
	Message string `json:"message"`
	// PrecisionUnchanged is true when a Precision was sent without a Value, so
	// Core kept the stored precision.
	PrecisionUnchanged bool `json:"precision_unchanged,omitempty"`
}

// DeleteBalanceMonitorResponse is returned when a balance monitor is deleted.
type DeleteBalanceMonitorResponse struct {
	Message string `json:"message"`
}

func (s *BalanceMonitorService) Create(data MonitorData) (*MonitorDataResp, *http.Response, error) {
	if err := ValidateMonitorTrigger(data.Trigger); err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest("balance-monitors", http.MethodPost, data)
	if err != nil {
		return nil, nil, err
	}

	monitorData := new(MonitorDataResp)
	resp, err := s.client.CallWithRetry(req, monitorData)
	if err != nil {
		return nil, resp, err
	}

	return monitorData, resp, nil
}

func (s *BalanceMonitorService) Get(monitorID string) (*MonitorDataResp, *http.Response, error) {
	u, err := endpointPath("balance-monitors/%s", monitorID)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(u, http.MethodGet, nil)
	if err != nil {
		return nil, nil, err
	}

	var resp MonitorDataResp
	httpResp, err := s.client.CallWithRetry(req, &resp)
	if err != nil {
		return nil, httpResp, err
	}

	return &resp, httpResp, nil
}

func (s *BalanceMonitorService) List() ([]MonitorDataResp, *http.Response, error) {
	req, err := s.client.NewRequest("balance-monitors", http.MethodGet, nil)
	if err != nil {
		return nil, nil, err
	}

	var monitorData []MonitorDataResp
	resp, err := s.client.CallWithRetry(req, &monitorData)
	if err != nil {
		return nil, resp, err
	}

	return monitorData, resp, nil
}

// ListByBalanceID returns monitors for a specific balance via
// GET /balance-monitors/balances/{balance_id}.
func (s *BalanceMonitorService) ListByBalanceID(balanceID string) ([]MonitorDataResp, *http.Response, error) {
	if err := ValidateBalanceID(balanceID); err != nil {
		return nil, nil, err
	}

	u, err := endpointPath("balance-monitors/balances/%s", balanceID)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(u, http.MethodGet, nil)
	if err != nil {
		return nil, nil, err
	}

	var monitorData []MonitorDataResp
	resp, err := s.client.CallWithRetry(req, &monitorData)
	if err != nil {
		return nil, resp, err
	}

	return monitorData, resp, nil
}

// Update sends a full MonitorData to PUT /balance-monitors/{id}.
//
// Deprecated: Core replies with a message, not the monitor, so the returned
// MonitorDataResp is empty, and every zero-valued field is sent. Use
// UpdatePartial, which sends only the fields you set.
func (s *BalanceMonitorService) Update(monitorID string, data MonitorData) (*MonitorDataResp, *http.Response, error) {
	if err := ValidateMonitorTrigger(data.Trigger); err != nil {
		return nil, nil, err
	}
	u, err := endpointPath("balance-monitors/%s", monitorID)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(u, http.MethodPut, data)
	if err != nil {
		return nil, nil, err
	}

	monitorData := new(MonitorDataResp)
	resp, err := s.client.CallWithRetry(req, monitorData)
	if err != nil {
		return nil, resp, err
	}

	return monitorData, resp, nil
}

// UpdatePartial changes a balance monitor via PUT /balance-monitors/{id},
// sending only the fields set on data. Fields left nil keep their stored
// values (Core 0.15.9+).
func (s *BalanceMonitorService) UpdatePartial(monitorID string, data UpdateMonitorRequest) (*UpdateBalanceMonitorResponse, *http.Response, error) {
	if err := ValidateMonitorID(monitorID); err != nil {
		return nil, nil, err
	}
	if err := ValidateMonitorTrigger(data.Trigger); err != nil {
		return nil, nil, err
	}

	u, err := endpointPath("balance-monitors/%s", monitorID)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(u, http.MethodPut, data)
	if err != nil {
		return nil, nil, err
	}

	updateResp := new(UpdateBalanceMonitorResponse)
	resp, err := s.client.CallWithRetry(req, updateResp)
	if err != nil {
		return nil, resp, err
	}

	return updateResp, resp, nil
}

// Delete removes a balance monitor by ID (Core 0.15.0+).
func (s *BalanceMonitorService) Delete(monitorID string) (*DeleteBalanceMonitorResponse, *http.Response, error) {
	if err := ValidateMonitorID(monitorID); err != nil {
		return nil, nil, err
	}

	u, err := endpointPath("balance-monitors/%s", monitorID)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(u, http.MethodDelete, nil)
	if err != nil {
		return nil, nil, err
	}

	deleteResp := new(DeleteBalanceMonitorResponse)
	resp, err := s.client.CallWithRetry(req, deleteResp)
	if err != nil {
		return nil, resp, err
	}

	return deleteResp, resp, nil
}

func NewBalanceMonitorService(client ClientInterface) *BalanceMonitorService {

	return &BalanceMonitorService{client: client}

}
