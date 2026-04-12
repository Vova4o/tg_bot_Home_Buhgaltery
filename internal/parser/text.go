package parser

import (
	"regexp"
	"strconv"
	"strings"
)

type TextExpense struct {
	Description string
	Amount      float64
}

// ExtractExpense extracts the description and amount from a text like "bread 50" or "taxi 430.50"
func ExtractExpense(text string) (*TextExpense, error) {
	// Normalize spaces
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// Regex to find amount at the end of string. Supports floats (e.g. 50, 50.50, 50,50)
	re := regexp.MustCompile(`^(.*)\s+([\d]+[.,]?[\d]*)\s*(?:rub|₽)?$`)
	matches := re.FindStringSubmatch(strings.ToLower(text))

	if len(matches) != 3 {
		// Maybe just the amount?
		reJustAmount := regexp.MustCompile(`^([\d]+[.,]?[\d]*)\s*(?:rub|₽)?$`)
		if matchesJustAmount := reJustAmount.FindStringSubmatch(strings.ToLower(text)); len(matchesJustAmount) == 2 {
			amountStr := strings.Replace(matchesJustAmount[1], ",", ".", 1)
			amount, err := strconv.ParseFloat(amountStr, 64)
			if err != nil {
				return nil, err
			}
			return &TextExpense{
				Description: "",
				Amount:      amount,
			}, nil
		}
		return nil, nil // Could not parse
	}

	description := strings.TrimSpace(matches[1])
	amountStr := strings.Replace(matches[2], ",", ".", 1)

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return nil, err
	}

	return &TextExpense{
		Description: description,
		Amount:      amount,
	}, nil
}
