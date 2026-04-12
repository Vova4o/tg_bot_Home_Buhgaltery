package ocr

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/otiai10/gosseract/v2"
)

type ParsedReceipt struct {
	RawText      string
	TotalAmount  float64
	MerchantName string
	PurchaseAt   time.Time
	Items        []ReceiptItem
}

type ReceiptItem struct {
	Name   string
	Amount float64
}

// PerformOCR reads an image file and returns the extracted text using Tesseract.
func PerformOCR(filePath string) (string, error) {
	client := gosseract.NewClient()
	defer client.Close()

	// Use both Russian and English
	client.SetLanguage("rus", "eng")
	client.SetImage(filePath)

	text, err := client.Text()
	if err != nil {
		return "", err
	}

	return text, nil
}

// ParseReceiptText takes the raw OCR text and attempts to extract relevant details.
func ParseReceiptText(text string) *ParsedReceipt {
	receipt := &ParsedReceipt{
		RawText: text,
		Items:   []ReceiptItem{},
	}

	lines := strings.Split(text, "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleanedLines = append(cleanedLines, trimmed)
		}
	}

	if len(cleanedLines) > 0 {
		// Heuristic: First line is often the merchant name
		receipt.MerchantName = cleanedLines[0]
	}

	// Regex for Totals
	totalRe := regexp.MustCompile(`(?i)(?:ИТОГ|СУММА|ИТОГО|TOTAL|SUM|AMOUNT|ИТОГ:|СУММА:|TOTAL:)\s*(\d+[.,]\d{2})`)
	// Fallback regex to just look for largest number if no total keyword found (simple heuristic, maybe unreliable)

	// Regex for Date (DD.MM.YYYY, YYYY-MM-DD, DD/MM/YYYY)
	dateRe := regexp.MustCompile(`(\d{2}[./-]\d{2}[./-]\d{4}|\d{4}[./-]\d{2}[./-]\d{2})`)

	// Regex for Item lines (some text followed by a price)
	// Example: "Bread 50.00", "Молоко 120,50"
	itemRe := regexp.MustCompile(`^(.*?)\s+(\d+[.,]\d{2})$`)

	var maxAmount float64

	for _, line := range cleanedLines {
		// Try to find Date
		if receipt.PurchaseAt.IsZero() {
			dateMatch := dateRe.FindString(line)
			if dateMatch != "" {
				// Normalize date separator to dash for simpler parsing, or just use custom formats
				d, err := parseDate(dateMatch)
				if err == nil {
					receipt.PurchaseAt = d
				}
			}
		}

		// Try to find Total explicitly
		totalMatch := totalRe.FindStringSubmatch(line)
		if len(totalMatch) > 1 {
			amountStr := strings.Replace(totalMatch[1], ",", ".", 1)
			amount, err := strconv.ParseFloat(amountStr, 64)
			if err == nil {
				receipt.TotalAmount = amount
			}
		} else {
			// Try to find items (excluding lines with "ИТОГ", etc)
			if !strings.Contains(strings.ToUpper(line), "ИТОГ") && !strings.Contains(strings.ToUpper(line), "TOTAL") {
				itemMatch := itemRe.FindStringSubmatch(line)
				if len(itemMatch) > 2 {
					name := strings.TrimSpace(itemMatch[1])
					amountStr := strings.Replace(itemMatch[2], ",", ".", 1)
					amount, err := strconv.ParseFloat(amountStr, 64)
					if err == nil && len(name) > 1 {
						receipt.Items = append(receipt.Items, ReceiptItem{
							Name:   name,
							Amount: amount,
						})
						if amount > maxAmount {
							maxAmount = amount
						}
					}
				}
			}
		}
	}

	// If no explicit total is found, fallback to the largest item amount, or sum of items if they seem like a list
	// This is a naive fallback. Real receipt parsing is complex.
	if receipt.TotalAmount == 0 {
		if maxAmount > 0 {
			log.Println("Fallback: using max amount found as total")
			receipt.TotalAmount = maxAmount
		}
	}

	// Default to now if no date found
	if receipt.PurchaseAt.IsZero() {
		receipt.PurchaseAt = time.Now()
	}

	return receipt
}

func parseDate(dateStr string) (time.Time, error) {
	dateStr = strings.ReplaceAll(dateStr, "/", ".")
	dateStr = strings.ReplaceAll(dateStr, "-", ".")

	layouts := []string{
		"02.01.2006",
		"2006.01.02",
	}

	var d time.Time
	var err error
	for _, layout := range layouts {
		d, err = time.Parse(layout, dateStr)
		if err == nil {
			return d, nil
		}
	}
	return time.Time{}, err
}
