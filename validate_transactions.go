package blnkgo

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"
)

const distributionSumEpsilon = 1e-9

func ValidateCreateTransacation(t CreateTransactionRequest) error {
	var sb strings.Builder
	sb.WriteString("validation error:")
	if t.Source != "" && len(t.Sources) > 0 {
		sb.WriteString("you can not use both Source and Sources")
		return errors.New(sb.String())
	}

	if t.Source == "" && len(t.Sources) == 0 {
		sb.WriteString("you must use either Source or Sources")
		return errors.New(sb.String())
	}

	if t.Destination != "" && len(t.Destinations) > 0 {
		sb.WriteString("you can not use both Destination and Destinations")
		return errors.New(sb.String())
	}

	if t.Destination == "" && len(t.Destinations) == 0 {
		sb.WriteString("you must use either Destination or Destinations")
		return errors.New(sb.String())
	}

	if t.Amount < 0 {
		sb.WriteString("you can not use a negative amount")
		return errors.New(sb.String())
	}

	if len(t.Sources) > 0 && shouldValidateSplitLegSums(t, t.Sources) {
		if err := validateSplitLegs(t.Sources, t, "source"); err != nil {
			return err
		}
	}

	if len(t.Destinations) > 0 && shouldValidateSplitLegSums(t, t.Destinations) {
		if err := validateSplitLegs(t.Destinations, t, "destination"); err != nil {
			return err
		}
	}

	return nil
}

func ValidateBulkCommitInflight(b BulkCommitInflightRequest) error {
	if len(b.Transactions) == 0 {
		return errors.New("validation error: transactions array cannot be empty")
	}
	if len(b.Transactions) > MaxBulkInflightItems {
		return fmt.Errorf("validation error: too many transactions; max is %d", MaxBulkInflightItems)
	}
	for i, tx := range b.Transactions {
		if tx.TransactionID == "" {
			return fmt.Errorf("validation error: transaction_id is required at index %d", i)
		}
	}
	return nil
}

func ValidateBulkVoidInflight(b BulkVoidInflightRequest) error {
	if len(b.TransactionIDs) == 0 {
		return errors.New("validation error: transaction_ids array cannot be empty")
	}
	if len(b.TransactionIDs) > MaxBulkInflightItems {
		return fmt.Errorf("validation error: too many transaction_ids; max is %d", MaxBulkInflightItems)
	}
	for i, id := range b.TransactionIDs {
		if id == "" {
			return fmt.Errorf("validation error: transaction_id is required at index %d", i)
		}
	}
	return nil
}

func ValidateCreateBulkTransaction(b CreateBulkTransactionRequest) error {
	if len(b.Transactions) == 0 {
		return errors.New("validation error: transactions array cannot be empty")
	}
	if len(b.Transactions) > MaxBulkCreateItems {
		return fmt.Errorf("validation error: too many transactions; max is %d", MaxBulkCreateItems)
	}

	refs := make(map[string]struct{}, len(b.Transactions))
	for i, tx := range b.Transactions {
		if err := ValidateCreateTransacation(tx); err != nil {
			return fmt.Errorf("transaction at index %d: %w", i, err)
		}
		if _, exists := refs[tx.Reference]; exists {
			return errors.New("validation error: all transactions must have unique references within the bulk request")
		}
		refs[tx.Reference] = struct{}{}
	}

	return nil
}

func ValidateRecoverQueue(r RecoverQueueRequest) error {
	if r.Threshold == "" {
		return nil
	}
	if _, err := time.ParseDuration(r.Threshold); err != nil {
		return fmt.Errorf("threshold must be a valid duration string (e.g. 5m, 1h)")
	}
	return nil
}

func ValidateRefundTransaction(r RefundTransactionRequest) error {
	// RefundTransactionRequest only exposes skip_queue; no extra field validation needed
	// beyond JSON types. Kept for parity with other transaction validators.
	return nil
}

func hasPreciseDistribution(leg Source) bool {
	return leg.PreciseDistribution != ""
}

func legsUsePreciseDistribution(legs []Source) bool {
	for _, leg := range legs {
		if hasPreciseDistribution(leg) {
			return true
		}
	}
	return false
}

// shouldValidateSplitLegSums preserves pre-#71 behavior: when precise_amount is set
// and no leg uses precise_distribution, Core remains the source of truth for
// classic distribution splits. Validate sums only for amount-based splits or when
// precise_distribution appears on any leg.
func shouldValidateSplitLegSums(t CreateTransactionRequest, legs []Source) bool {
	if legsUsePreciseDistribution(legs) {
		return true
	}
	if t.PreciseAmount != nil && t.PreciseAmount.Cmp(big.NewInt(0)) != 0 {
		return false
	}
	return true
}

func usesPreciseIntegerArithmetic(t CreateTransactionRequest) bool {
	for _, leg := range t.Sources {
		if hasPreciseDistribution(leg) {
			return true
		}
	}
	for _, leg := range t.Destinations {
		if hasPreciseDistribution(leg) {
			return true
		}
	}
	if t.PreciseAmount != nil && t.Amount == 0 {
		return true
	}
	return false
}

func parsePreciseInteger(raw string) (*big.Int, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed != raw {
		return nil, false
	}
	n := new(big.Int)
	if _, ok := n.SetString(trimmed, 10); !ok || n.Sign() < 0 {
		return nil, false
	}
	return n, true
}

func legUsesDecimalDistribution(leg Source) bool {
	if leg.Distribution == "" {
		return false
	}
	if leg.Distribution.IsPercentage() {
		pct := leg.Distribution.ToPercentage()
		return pct != math.Trunc(pct)
	}
	if leg.Distribution.IsNumber() {
		num := leg.Distribution.ToNumber()
		return num != math.Trunc(num)
	}
	return false
}

type splitLegTotal struct {
	bigInt *big.Int
	float  float64
	mode   string
}

func resolveSplitLegTotal(t CreateTransactionRequest) (splitLegTotal, error) {
	var zero splitLegTotal
	hasPreciseAmount := t.PreciseAmount != nil

	if t.Amount == 0 && !hasPreciseAmount {
		return zero, errors.New("validation error: amount or precise_amount is required for split legs")
	}

	if usesPreciseIntegerArithmetic(t) {
		if t.Amount != 0 {
			if t.Amount != math.Trunc(t.Amount) {
				return zero, errors.New("validation error: amount must be a whole number when split legs use precise_distribution")
			}
			return splitLegTotal{bigInt: big.NewInt(int64(t.Amount)), mode: "bigint"}, nil
		}
		if hasPreciseAmount {
			return splitLegTotal{bigInt: new(big.Int).Set(t.PreciseAmount), mode: "bigint"}, nil
		}
	}

	if t.Amount != 0 {
		return splitLegTotal{float: t.Amount, mode: "float"}, nil
	}

	if hasPreciseAmount {
		if t.PreciseAmount.Sign() < 0 {
			return zero, errors.New("validation error: precise_amount must be non-negative")
		}
		maxSafe := big.NewInt(1<<53 - 1)
		if t.PreciseAmount.Cmp(maxSafe) <= 0 {
			f, _ := new(big.Float).SetInt(t.PreciseAmount).Float64()
			return splitLegTotal{float: f, mode: "float"}, nil
		}
		return splitLegTotal{bigInt: new(big.Int).Set(t.PreciseAmount), mode: "bigint"}, nil
	}

	return zero, errors.New("validation error: amount or precise_amount is required for split legs")
}

func validateSplitLegs(legs []Source, t CreateTransactionRequest, legLabel string) error {
	total, err := resolveSplitLegTotal(t)
	if err != nil {
		return err
	}

	for _, leg := range legs {
		if strings.TrimSpace(leg.Identifier) == "" {
			return fmt.Errorf("validation error: each %s leg must include a valid identifier", legLabel)
		}

		hasDistribution := leg.Distribution != ""
		hasPrecise := hasPreciseDistribution(leg)
		if !hasDistribution && !hasPrecise {
			return fmt.Errorf("validation error: each %s leg must include either 'distribution' or 'precise_distribution'", legLabel)
		}
	}

	if total.mode == "bigint" {
		return validateSplitLegsBigInt(legs, total.bigInt)
	}
	return validateSplitLegsFloat(legs, total.float)
}

func validateSplitLegsBigInt(legs []Source, total *big.Int) error {
	hasDecimalDistribution := false
	for _, leg := range legs {
		if legUsesDecimalDistribution(leg) {
			hasDecimalDistribution = true
			break
		}
	}

	maxSafe := big.NewInt(1 << 53)
	if hasDecimalDistribution && total.Cmp(maxSafe) > 0 {
		return errors.New("validation error: decimal distribution values are not supported with precise amounts beyond Number.MAX_SAFE_INTEGER")
	}

	if hasDecimalDistribution && total.Cmp(maxSafe) <= 0 {
		f, _ := new(big.Float).SetInt(total).Float64()
		return validateSplitLegsFloat(legs, f)
	}

	sum := big.NewInt(0)
	hasLeft := false

	for _, leg := range legs {
		if hasPreciseDistribution(leg) {
			preciseValue, ok := parsePreciseInteger(leg.PreciseDistribution)
			if !ok {
				return fmt.Errorf("validation error: invalid precise_distribution for leg: %s", leg.Identifier)
			}
			sum.Add(sum, preciseValue)
			continue
		}

		distribution := leg.Distribution
		if !distribution.IsValid() {
			return fmt.Errorf("validation error: invalid distribution type for leg: %s", leg.Identifier)
		}

		switch {
		case distribution.IsPercentage():
			percentage := distribution.ToPercentage()
			if percentage < 0 || percentage > 100 {
				return fmt.Errorf("validation error: invalid percentage value in leg: %s", leg.Identifier)
			}
			pct := big.NewInt(int64(percentage))
			part := new(big.Int).Mul(total, pct)
			part.Div(part, big.NewInt(100))
			sum.Add(sum, part)

		case distribution.IsLeft():
			if hasLeft {
				return errors.New("validation error: multiple 'left' distribution types are not allowed")
			}
			hasLeft = true

		case distribution.IsNumber():
			fixedAmount := distribution.ToNumber()
			if fixedAmount < 0 || fixedAmount != math.Trunc(fixedAmount) {
				return fmt.Errorf("validation error: invalid distribution type for leg: %s", leg.Identifier)
			}
			sum.Add(sum, big.NewInt(int64(fixedAmount)))

		default:
			return fmt.Errorf("validation error: invalid distribution type for leg: %s", leg.Identifier)
		}
	}

	if hasLeft {
		remaining := new(big.Int).Sub(total, sum)
		if remaining.Sign() < 0 {
			return errors.New("validation error: total distribution exceeds the specified amount")
		}
	} else if sum.Cmp(total) != 0 {
		return fmt.Errorf("validation error: total distribution sum (%s) does not equal the specified amount (%s)", sum.String(), total.String())
	}

	return nil
}

func validateSplitLegsFloat(legs []Source, amount float64) error {
	sum := 0.0
	hasLeft := false

	for _, leg := range legs {
		if hasPreciseDistribution(leg) {
			preciseValue, ok := parsePreciseInteger(leg.PreciseDistribution)
			if !ok {
				return fmt.Errorf("validation error: invalid precise_distribution for leg: %s", leg.Identifier)
			}
			maxSafe := big.NewInt(1 << 53)
			if preciseValue.Cmp(maxSafe) > 0 {
				return fmt.Errorf("validation error: invalid precise_distribution for leg: %s", leg.Identifier)
			}
			f, _ := new(big.Float).SetInt(preciseValue).Float64()
			sum += f
			continue
		}

		distribution := leg.Distribution
		if !distribution.IsValid() {
			return fmt.Errorf("validation error: invalid distribution: %s", distribution)
		}

		switch {
		case distribution.IsPercentage():
			percentage := distribution.ToPercentage()
			if percentage < 0 || percentage > 100 {
				return fmt.Errorf("validation error: invalid percentage value in leg: %s", leg.Identifier)
			}
			sum += (percentage / 100) * amount

		case distribution.IsNumber():
			number := distribution.ToNumber()
			if number < 0 {
				return fmt.Errorf("validation error: invalid distribution in source: %s", leg.Identifier)
			}
			sum += number

		case distribution.IsLeft():
			if hasLeft {
				return errors.New("validation error: you cannot use left distribution more than once")
			}
			hasLeft = true

		default:
			return fmt.Errorf("validation error: unknown distribution type in source: %s", leg.Identifier)
		}
	}

	if hasLeft {
		remaining := amount - sum
		if remaining < -distributionSumEpsilon {
			return errors.New("validation error: total distribution exceeds the specified amount")
		}
	} else if !distributionTotalsApproximatelyEqual(sum, amount) {
		return errors.New("validation error: total amount of sources must be equal to the amount")
	}

	return nil
}

func distributionTotalsApproximatelyEqual(sum, amount float64) bool {
	return math.Abs(sum-amount) <= distributionSumEpsilon
}
