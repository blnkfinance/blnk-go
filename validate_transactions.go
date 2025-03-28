package blnkgo

import (
	"errors"
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

	if len(t.Sources) > 0 && t.PreciseAmount.Cmp(big.NewInt(0)) == 0 {
		err := validateSources(t.Sources, t.Amount, &sb)
		if err != nil {
			return err
		}
	}

	if len(t.Destinations) > 0 && t.PreciseAmount.Cmp(big.NewInt(0)) == 0 {
		err := validateSources(t.Destinations, t.Amount, &sb)
		if err != nil {
			return err
		}
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
