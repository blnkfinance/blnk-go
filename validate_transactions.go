package blnkgo

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

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

	if len(t.Sources) > 0 && (t.PreciseAmount == nil || t.PreciseAmount.Cmp(big.NewInt(0)) == 0) {
		err := validateSources(t.Sources, t.Amount, &sb)
		if err != nil {
			return err
		}
	}

	if len(t.Destinations) > 0 && (t.PreciseAmount == nil || t.PreciseAmount.Cmp(big.NewInt(0)) == 0) {
		err := validateSources(t.Destinations, t.Amount, &sb)
		if err != nil {
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

func validateSources(sources []Source, amount float64, sb *strings.Builder) error {
	//total amount of sources  must be equal to the amount
	total := 0.0
	hasLeft := false
	for _, source := range sources {
		distribution := source.Distribution
		//check if the distribution is valid
		isValid := distribution.IsValid()
		if !isValid {
			sb.WriteString("invalid distribution: " + string(distribution))
			return errors.New(sb.String())
		}

		switch {
		case distribution.IsPercentage():
			// Get float value from percentage
			percentage := distribution.ToPercentage()
			v := (percentage / 100) * amount
			if v < 0 {
				sb.WriteString("invalid distribution in source: " + source.Identifier)
				return errors.New(sb.String())
			}
			total += v

		case distribution.IsNumber():
			// Get float value from number
			number := distribution.ToNumber()
			if number < 0 {
				sb.WriteString("invalid distribution in source: " + source.Identifier)
				return errors.New(sb.String())
			}
			total += number

		case distribution.IsLeft():
			// Ensure "left" distribution is used only once
			if hasLeft {
				sb.WriteString("you cannot use left distribution more than once")
				return errors.New(sb.String())
			}
			hasLeft = true

		default:
			sb.WriteString("unknown distribution type in source: " + source.Identifier)
			// Handle invalid or unrecognized distribution
			return errors.New(sb.String())
		}
	}

	// If "left" distribution is used, calculate its value and add to total
	if hasLeft {
		left := amount - total
		if left < 0 {
			sb.WriteString("total amount of sources exceeds the amount")
			return errors.New(sb.String())
		}
		total += left
	}

	if total != amount {
		sb.WriteString("total amount of sources must be equal to the amount")
		return errors.New(sb.String())
	}

	return nil
}
