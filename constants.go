package blnkgo

import (
	"regexp"
	"strconv"
)

type Distribution string

var (
	percentageRegex = regexp.MustCompile(`^\d+(\.\d+)?%$`)
	numberRegex     = regexp.MustCompile(`^\d+(\.\d+)?$`)
)

func (d Distribution) IsValid() bool {
	return percentageRegex.MatchString(string(d)) ||
		numberRegex.MatchString(string(d)) ||
		d == "left"
}

func (d Distribution) IsPercentage() bool {
	return percentageRegex.MatchString(string(d))
}

func (d Distribution) IsNumber() bool {
	return numberRegex.MatchString(string(d))
}

func (d Distribution) IsLeft() bool {
	return d == "left"
}

// Convert percentage string to a float64
func (d Distribution) ToPercentage() float64 {
	if d.IsPercentage() {
		// Remove the '%' character and parse the remaining string
		percentage, err := strconv.ParseFloat(string(d[:len(d)-1]), 64)
		if err == nil {
			return percentage
		}
	}
	return 0
}

// Convert number string to a float64
func (d Distribution) ToNumber() float64 {
	if d.IsNumber() {
		number, err := strconv.ParseFloat(string(d), 64)
		if err == nil {
			return number
		}
	}
	return 0
}

// PryTransactionStatus represents the transaction status.
type PryTransactionStatus string

const (
	PryTransactionStatusQueued   PryTransactionStatus = "QUEUED"
	PryTransactionStatusApplied  PryTransactionStatus = "APPLIED"
	PryTransactionStatusRejected PryTransactionStatus = "REJECTED"
	PryTransactionStatusCommit   PryTransactionStatus = "COMMIT"
	PryTransactionStatusVoid     PryTransactionStatus = "VOID"
	PryTransactionStatusInFlight PryTransactionStatus = "INFLIGHT"
	PryTransactionStatusExpired  PryTransactionStatus = "EXPIRED"
)

// InflightStatus represents the status of inflight transactions.
type InflightStatus string

const (
	InflightStatusCommit InflightStatus = "commit"
	InflightStatusVoid   InflightStatus = "void"
)

type IdentityType string

const (
	Individual   IdentityType = "individual"
	Organization IdentityType = "organization"
)

// MonitorConditionOperators represents the supported operators for a monitor condition.
type MonitorConditionOperators string

const (
	OperatorGreaterThan        MonitorConditionOperators = ">"
	OperatorLessThan           MonitorConditionOperators = "<"
	OperatorEqualTo            MonitorConditionOperators = "="
	OperatorNotEqualTo         MonitorConditionOperators = "!="
	OperatorGreaterThanOrEqual MonitorConditionOperators = ">="
	OperatorLessThanOrEqual    MonitorConditionOperators = "<="
)

type ResourceType string

const (
	Ledgers      ResourceType = "ledgers"
	Transactions ResourceType = "transactions"
	Balances     ResourceType = "balances"
	Identities   ResourceType = "identities"
)

type CriteriaField string

const (
	CriteriaFieldAmount      CriteriaField = "amount"
	CriteriaFieldCurrency    CriteriaField = "currency"
	CriteriaFieldReference   CriteriaField = "reference"
	CriteriaFieldDescription CriteriaField = "description"
	CriteriaFieldDate        CriteriaField = "date"
)

type ReconciliationOperator string

const (
	ReconciliationOperatorEquals      ReconciliationOperator = "equals"
	ReconciliationOperatorGreaterThan ReconciliationOperator = "greater_than"
	ReconciliationOperatorLessThan    ReconciliationOperator = "less_than"
	ReconciliationOperatorContains    ReconciliationOperator = "contains"
)

// Strategy represents the allowed reconciliation strategies.
type ReconciliationStrategy string

const (
	ReconciliationStrategyOneToOne  ReconciliationStrategy = "one_to_one"
	ReconciliationStrategyOneToMany ReconciliationStrategy = "one_to_many"
	ReconciliationStrategyManyToOne ReconciliationStrategy = "many_to_one"
)
