package db

import (
	"database/sql"
	"time"

	"expense-bot/internal/models"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Repository struct {
	DB *sqlx.DB
}

func NewRepository(dbURL string) (*Repository, error) {
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		return nil, err
	}
	return &Repository{DB: db}, nil
}

func (r *Repository) GetOrCreateUser(telegramUserID int64, displayName string) (*models.User, error) {
	var user models.User
	err := r.DB.Get(&user, "SELECT * FROM users WHERE telegram_user_id = $1", telegramUserID)
	if err == sql.ErrNoRows {
		// Create new user
		err = r.DB.QueryRowx("INSERT INTO users (telegram_user_id, display_name) VALUES ($1, $2) RETURNING id, telegram_user_id, display_name, created_at",
			telegramUserID, displayName).StructScan(&user)
		if err != nil {
			return nil, err
		}
		return &user, nil
	} else if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetCategoryByName(name string) (*models.Category, error) {
	var category models.Category
	err := r.DB.Get(&category, "SELECT * FROM categories WHERE name = $1", name)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *Repository) AddExpense(expense *models.Expense) error {
	query := `
		INSERT INTO expenses (user_id, category_id, source_type, description, amount, expense_at)
		VALUES (:user_id, :category_id, :source_type, :description, :amount, :expense_at)
		RETURNING id, created_at`
	rows, err := r.DB.NamedQuery(query, expense)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.StructScan(expense)
	}
	return err
}

func (r *Repository) SaveReceipt(receipt *models.Receipt) error {
	query := `
		INSERT INTO receipts (user_id, telegram_message_id, file_path, ocr_text, merchant_name, purchase_at, total_amount, currency)
		VALUES (:user_id, :telegram_message_id, :file_path, :ocr_text, :merchant_name, :purchase_at, :total_amount, :currency)
		RETURNING id, created_at`
	rows, err := r.DB.NamedQuery(query, receipt)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.StructScan(receipt)
	}
	return err
}

func (r *Repository) SaveReceiptItem(item *models.ReceiptItem) error {
	query := `
		INSERT INTO receipt_items (receipt_id, item_name, amount, category_id)
		VALUES (:receipt_id, :item_name, :amount, :category_id)
		RETURNING id`
	rows, err := r.DB.NamedQuery(query, item)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.StructScan(item)
	}
	return err
}

// Time helper to get bounds for Moscow timezone
func getBounds(period string) (time.Time, time.Time, error) {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	now := time.Now().In(loc)

	var start time.Time
	var end time.Time

	switch period {
	case "day":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		end = start.AddDate(0, 0, 1)
	case "week":
		// Week starts on Monday
		offset := int(time.Monday - now.Weekday())
		if offset > 0 {
			offset = -6
		}
		start = time.Date(now.Year(), now.Month(), now.Day()+offset, 0, 0, 0, 0, loc)
		end = start.AddDate(0, 0, 7)
	case "month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		end = start.AddDate(0, 1, 0)
	}
	return start, end, nil
}

func (r *Repository) GetExpensesReport(userID int, period string) ([]models.ExpenseReportRow, error) {
	start, end, err := getBounds(period)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT c.name as category_name, SUM(e.amount) as total_amount
		FROM expenses e
		JOIN categories c ON e.category_id = c.id
		WHERE e.user_id = $1 AND e.expense_at >= $2 AND e.expense_at < $3
		GROUP BY c.name
		ORDER BY total_amount DESC`

	var rows []models.ExpenseReportRow
	err = r.DB.Select(&rows, query, userID, start, end)
	return rows, err
}

func (r *Repository) GetCategoryExpensesReport(userID int, categoryName string, period string) (float64, error) {
	start, end, err := getBounds(period)
	if err != nil {
		return 0, err
	}

	query := `
		SELECT COALESCE(SUM(e.amount), 0)
		FROM expenses e
		JOIN categories c ON e.category_id = c.id
		WHERE e.user_id = $1 AND c.name = $2 AND e.expense_at >= $3 AND e.expense_at < $4`

	var total float64
	err = r.DB.Get(&total, query, userID, categoryName, start, end)
	return total, err
}

func (r *Repository) GetUserExpensesReport(userID int, period string) (float64, error) {
	start, end, err := getBounds(period)
	if err != nil {
		return 0, err
	}

	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM expenses
		WHERE user_id = $1 AND expense_at >= $2 AND expense_at < $3`

	var total float64
	err = r.DB.Get(&total, query, userID, start, end)
	return total, err
}

func (r *Repository) GetLastExpenses(userID int, limit int) ([]models.Expense, error) {
	query := `
		SELECT * FROM expenses
		WHERE user_id = $1
		ORDER BY expense_at DESC
		LIMIT $2`
	var expenses []models.Expense
	err := r.DB.Select(&expenses, query, userID, limit)
	return expenses, err
}

func (r *Repository) SaveReceiptItems(items []*models.ReceiptItem) error {
	if len(items) == 0 {
		return nil
	}
	query := `
		INSERT INTO receipt_items (receipt_id, item_name, amount, category_id)
		VALUES (:receipt_id, :item_name, :amount, :category_id)
		RETURNING id`

	// NamedQuery works with slice of structs/pointers in sqlx >= 1.3.0
	// But it might not return correctly multiple rows if we use it directly without iterating or using NamedQuery over slice.
	// Actually sqlx.NamedQuery supports bulk insert if we pass a slice, but it returns multiple rows.

	rows, err := r.DB.NamedQuery(query, items)
	if err != nil {
		return err
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		if i < len(items) {
			err = rows.StructScan(items[i])
			if err != nil {
				return err
			}
			i++
		}
	}
	return rows.Err()
}

func (r *Repository) AddExpenses(expenses []*models.Expense) error {
	if len(expenses) == 0 {
		return nil
	}
	query := `
		INSERT INTO expenses (user_id, category_id, source_type, description, amount, expense_at)
		VALUES (:user_id, :category_id, :source_type, :description, :amount, :expense_at)
		RETURNING id, created_at`

	rows, err := r.DB.NamedQuery(query, expenses)
	if err != nil {
		return err
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		if i < len(expenses) {
			err = rows.StructScan(expenses[i])
			if err != nil {
				return err
			}
			i++
		}
	}
	return rows.Err()
}
