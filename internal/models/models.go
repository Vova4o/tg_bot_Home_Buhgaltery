package models

import (
	"time"
)

type User struct {
	ID             int       `db:"id"`
	TelegramUserID int64     `db:"telegram_user_id"`
	DisplayName    string    `db:"display_name"`
	CreatedAt      time.Time `db:"created_at"`
}

type Category struct {
	ID       int    `db:"id"`
	Name     string `db:"name"`
	ParentID *int   `db:"parent_id"`
}

type Expense struct {
	ID          int       `db:"id"`
	UserID      int       `db:"user_id"`
	CategoryID  int       `db:"category_id"`
	SourceType  string    `db:"source_type"`
	Description string    `db:"description"`
	Amount      float64   `db:"amount"`
	ExpenseAt   time.Time `db:"expense_at"`
	CreatedAt   time.Time `db:"created_at"`
}

type Receipt struct {
	ID                int       `db:"id"`
	UserID            int       `db:"user_id"`
	TelegramMessageID int64     `db:"telegram_message_id"`
	FilePath          string    `db:"file_path"`
	OCRText           string    `db:"ocr_text"`
	MerchantName      string    `db:"merchant_name"`
	PurchaseAt        time.Time `db:"purchase_at"`
	TotalAmount       float64   `db:"total_amount"`
	Currency          string    `db:"currency"`
	CreatedAt         time.Time `db:"created_at"`
}

type ReceiptItem struct {
	ID         int     `db:"id"`
	ReceiptID  int     `db:"receipt_id"`
	ItemName   string  `db:"item_name"`
	Amount     float64 `db:"amount"`
	CategoryID int     `db:"category_id"`
}

type ExpenseReportRow struct {
	CategoryName string  `db:"category_name"`
	TotalAmount  float64 `db:"total_amount"`
}
