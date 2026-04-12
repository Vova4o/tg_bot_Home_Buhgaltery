package parser

import (
	"testing"
)

func TestExtractExpense(t *testing.T) {
	tests := []struct {
		input       string
		description string
		amount      float64
		hasErr      bool
		nilExpected bool
	}{
		{"bread 50", "bread", 50, false, false},
		{"хлеб 50", "хлеб", 50, false, false},
		{"taxi 430", "taxi", 430, false, false},
		{"coffee 220.50", "coffee", 220.50, false, false},
		{"мясо 1200,50", "мясо", 1200.50, false, false},
		{"сигареты cigarettes 290", "сигареты cigarettes", 290, false, false},
		{"   bread   50  ", "bread", 50, false, false},
		{"bread 50 rub", "bread", 50, false, false},
		{"bread 50₽", "bread", 50, false, false},
		{"not an expense", "", 0, false, true},
	}

	for _, tc := range tests {
		res, err := ExtractExpense(tc.input)
		if tc.hasErr && err == nil {
			t.Errorf("Expected error for %q, got none", tc.input)
		}
		if !tc.hasErr && err != nil {
			t.Errorf("Unexpected error for %q: %v", tc.input, err)
		}
		if tc.nilExpected && res != nil {
			t.Errorf("Expected nil result for %q, got %v", tc.input, res)
		}
		if !tc.nilExpected && res == nil && !tc.hasErr {
			t.Errorf("Expected result for %q, got nil", tc.input)
		}

		if res != nil {
			if res.Description != tc.description {
				t.Errorf("For %q, expected description %q, got %q", tc.input, tc.description, res.Description)
			}
			if res.Amount != tc.amount {
				t.Errorf("For %q, expected amount %v, got %v", tc.input, tc.amount, res.Amount)
			}
		}
	}
}

func TestDetectCategory(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"bread", "groceries"},
		{"хлеб", "groceries"},
		{"taxi", "transport"},
		{"кофе", "cafe"},
		{"сигареты cigarettes", "cigarettes"}, // Matches first word
		{"random thing", "other"},
		{"купил молоко", "groceries"},
	}

	for _, tc := range tests {
		res := DetectCategory(tc.input)
		if res != tc.expected {
			t.Errorf("For %q, expected %q, got %q", tc.input, tc.expected, res)
		}
	}
}
