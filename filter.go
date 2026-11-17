package blnkgo

import (
	"encoding/json"
	"fmt"
)

type Operator string

const (
	OpEqual              Operator = "eq"
	OpNotEqual           Operator = "ne"
	OpGreaterThan        Operator = "gt"
	OpGreaterThanOrEqual Operator = "gte"
	OpLessThan           Operator = "lt"
	OpLessThanOrEqual    Operator = "lte"
	OpIn                 Operator = "in"
	OpBetween            Operator = "between"
	OpLike               Operator = "like"
	OpILike              Operator = "ilike"
	OpIsNull             Operator = "isnull"
	OpIsNotNull          Operator = "isnotnull"
)

// LogicalOperator joins the filters in a FilterParams.
type LogicalOperator string

const (
	// LogicalAnd matches records that satisfy every filter (Core's default).
	LogicalAnd LogicalOperator = "and"
	// LogicalOr matches records that satisfy any filter.
	LogicalOr LogicalOperator = "or"
)

type FilterParams struct {
	Filters []Filter `json:"filters"`
	// LogicalOperator is LogicalAnd or LogicalOr. Empty uses Core's default (and).
	LogicalOperator LogicalOperator `json:"logical_operator,omitempty"`
	Limit           int             `json:"limit,omitempty"`
	Offset          int             `json:"offset,omitempty"`
	SortBy          string          `json:"sort_by,omitempty"`
	SortOrder       string          `json:"sort_order,omitempty"`
	IncludeCount    bool            `json:"include_count,omitempty"`
}

type Filter struct {
	Field    string        `json:"field"`
	Operator Operator      `json:"operator"`
	Value    interface{}   `json:"value,omitempty"`
	Values   []interface{} `json:"values,omitempty"`
}

type FilterResponse struct {
	Data       interface{} `json:"data"`
	TotalCount *int64      `json:"total_count,omitempty"`
}

func (f *FilterResponse) UnmarshalJSON(data []byte) error {
	type alias FilterResponse
	var aux alias
	if err := json.Unmarshal(data, &aux); err == nil {
		f.Data = aux.Data
		f.TotalCount = aux.TotalCount
		return nil
	}

	var slice []interface{}
	if err := json.Unmarshal(data, &slice); err != nil {
		return fmt.Errorf("failed to decode FilterResponse: %w", err)
	}

	f.Data = slice
	f.TotalCount = nil
	return nil
}
