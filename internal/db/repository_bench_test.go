package db

import (
	"expense-bot/internal/models"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func setupMockRepo(b *testing.B) (*Repository, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("failed to open sqlmock: %v", err)
	}
	db := sqlx.NewDb(mockDB, "sqlmock")
	return &Repository{DB: db}, mock
}

// Benchmark the current N+1 behavior
func BenchmarkSaveReceiptItem_NPlus1(b *testing.B) {
	repo, mock := setupMockRepo(b)
	defer repo.DB.Close()

	items := make([]*models.ReceiptItem, 10)
	for i := 0; i < 10; i++ {
		items[i] = &models.ReceiptItem{
			ReceiptID:  1,
			ItemName:   "Test Item",
			Amount:     10.5,
			CategoryID: 2,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, item := range items {
			mock.ExpectQuery(`INSERT INTO receipt_items`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			repo.SaveReceiptItem(item)
		}
	}
}

// Placeholder for bulk insert benchmark
func BenchmarkSaveReceiptItem_Bulk(b *testing.B) {
	repo, mock := setupMockRepo(b)
	defer repo.DB.Close()

	items := make([]*models.ReceiptItem, 10)
	for i := 0; i < 10; i++ {
		items[i] = &models.ReceiptItem{
			ReceiptID:  1,
			ItemName:   "Test Item",
			Amount:     10.5,
			CategoryID: 2,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectQuery(`INSERT INTO receipt_items`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2).AddRow(3).AddRow(4).AddRow(5).AddRow(6).AddRow(7).AddRow(8).AddRow(9).AddRow(10))
		repo.SaveReceiptItems(items) // Call the bulk insert method once
	}
}

func BenchmarkSaveExpense_NPlus1(b *testing.B) {
	repo, mock := setupMockRepo(b)
	defer repo.DB.Close()

	expenses := make([]*models.Expense, 10)
	for i := 0; i < 10; i++ {
		expenses[i] = &models.Expense{
			UserID:      1,
			CategoryID:  2,
			SourceType:  "receipt",
			Description: "Test Expense",
			Amount:      10.5,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, exp := range expenses {
			mock.ExpectQuery(`INSERT INTO expenses`).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, time.Now()))
			repo.AddExpense(exp)
		}
	}
}

func BenchmarkSaveExpense_Bulk(b *testing.B) {
	repo, mock := setupMockRepo(b)
	defer repo.DB.Close()

	expenses := make([]*models.Expense, 10)
	for i := 0; i < 10; i++ {
		expenses[i] = &models.Expense{
			UserID:      1,
			CategoryID:  2,
			SourceType:  "receipt",
			Description: "Test Expense",
			Amount:      10.5,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectQuery(`INSERT INTO expenses`).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, time.Now()))
		repo.AddExpenses(expenses)
	}
}
