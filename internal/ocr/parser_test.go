package ocr

import (
	"testing"
	"time"
)

func TestParseReceiptText(t *testing.T) {
	rawText := `ООО "Ромашка"
Добро пожаловать
12.04.2023 15:30
Хлеб черный 50.00
Молоко 120,50
Кофе 300.00
ИТОГ 470.50
Спасибо за покупку!`

	receipt := ParseReceiptText(rawText)

	if receipt.MerchantName != "ООО \"Ромашка\"" {
		t.Errorf("Expected Merchant 'ООО \"Ромашка\"', got %q", receipt.MerchantName)
	}

	expectedDate, _ := time.Parse("02.01.2006", "12.04.2023")
	if !receipt.PurchaseAt.Equal(expectedDate) {
		t.Errorf("Expected Date %v, got %v", expectedDate, receipt.PurchaseAt)
	}

	if receipt.TotalAmount != 470.50 {
		t.Errorf("Expected TotalAmount 470.50, got %f", receipt.TotalAmount)
	}

	if len(receipt.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(receipt.Items))
	} else {
		if receipt.Items[0].Name != "Хлеб черный" || receipt.Items[0].Amount != 50.00 {
			t.Errorf("First item mismatch: %+v", receipt.Items[0])
		}
		if receipt.Items[1].Name != "Молоко" || receipt.Items[1].Amount != 120.50 {
			t.Errorf("Second item mismatch: %+v", receipt.Items[1])
		}
	}
}

func TestParseReceiptText_English(t *testing.T) {
	rawText := `WALMART
Welcome
2023-10-05 12:00
Apples 2.50
Bread 1.20
TOTAL 3.70
`
	receipt := ParseReceiptText(rawText)

	if receipt.MerchantName != "WALMART" {
		t.Errorf("Expected Merchant 'WALMART', got %q", receipt.MerchantName)
	}

	expectedDate, _ := time.Parse("2006.01.02", "2023.10.05")
	if !receipt.PurchaseAt.Equal(expectedDate) {
		t.Errorf("Expected Date %v, got %v", expectedDate, receipt.PurchaseAt)
	}

	if receipt.TotalAmount != 3.70 {
		t.Errorf("Expected TotalAmount 3.70, got %f", receipt.TotalAmount)
	}
}
