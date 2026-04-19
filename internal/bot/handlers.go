package bot

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"expense-bot/internal/models"
	"expense-bot/internal/ocr"
	"expense-bot/internal/parser"

	"gopkg.in/telebot.v4"
)

func (h *BotHandler) RegisterHandlers() {
	h.Bot.Handle(telebot.OnText, h.HandleText)
	h.Bot.Handle(telebot.OnPhoto, h.HandlePhoto)
	h.Bot.Handle("/day", h.HandleReportDay)
	h.Bot.Handle("/week", h.HandleReportWeek)
	h.Bot.Handle("/month", h.HandleReportMonth)
	h.Bot.Handle("/category", h.HandleReportCategory)
	h.Bot.Handle("/byuser", h.HandleReportUser)
	h.Bot.Handle("/last", h.HandleLast)
	h.Bot.Handle("/help", h.HandleHelp)
}

func (h *BotHandler) HandleText(c telebot.Context) error {
	text := c.Text()

	expenseParsed, err := parser.ExtractExpense(text)
	if err != nil || expenseParsed == nil {
		return c.Send("Не удалось распознать трату. Формат: <описание> <сумма>")
	}

	user, err := h.Repo.GetOrCreateUser(c.Sender().ID, c.Sender().Username)
	if err != nil {
		log.Println("Error getting user:", err)
		return c.Send("Ошибка базы данных")
	}

	categoryName := parser.DetectCategory(expenseParsed.Description)
	category, err := h.Repo.GetCategoryByName(categoryName)
	if err != nil {
		log.Println("Category not found:", categoryName, err)
		// Default to 'other' if the specific category isn't found
		c2, err2 := h.Repo.GetCategoryByName("other")
		if err2 == nil {
			category = c2
		} else {
			return c.Send("Ошибка: категория 'other' не найдена в базе данных")
		}
	}

	expense := &models.Expense{
		UserID:      user.ID,
		CategoryID:  category.ID,
		SourceType:  "text",
		Description: expenseParsed.Description,
		Amount:      expenseParsed.Amount,
		ExpenseAt:   time.Now(),
	}

	err = h.Repo.AddExpense(expense)
	if err != nil {
		log.Println("Error adding expense:", err)
		return c.Send("Ошибка сохранения траты")
	}

	catRus := parser.GetCategoryInRussian(category.Name)
	return c.Send(fmt.Sprintf("Добавлено: %s - %.2f RUB, категория: %s", expense.Description, expense.Amount, catRus))
}

func (h *BotHandler) HandlePhoto(c telebot.Context) error {
	photo := c.Message().Photo
	if photo == nil {
		return c.Send("Фото не найдено")
	}

	user, err := h.Repo.GetOrCreateUser(c.Sender().ID, c.Sender().Username)
	if err != nil {
		return c.Send("Ошибка пользователя")
	}

	file, err := h.Bot.FileByID(photo.FileID)
	if err != nil {
		return c.Send("Не удалось получить файл")
	}

	// Create tmp dir if not exists
	err = os.MkdirAll("downloads", os.ModePerm)
	if err != nil {
		return err
	}

	filePath := filepath.Join("downloads", fmt.Sprintf("%s.jpg", photo.FileID))
	err = h.Bot.Download(&file, filePath)
	if err != nil {
		return c.Send("Не удалось скачать фото")
	}
	defer os.Remove(filePath)

	rawText, err := ocr.PerformOCR(filePath)
	if err != nil {
		log.Println("OCR Error:", err)
		return c.Send("Ошибка распознавания чека")
	}

	parsedReceipt := ocr.ParseReceiptText(rawText)

	// Save Receipt
	receipt := &models.Receipt{
		UserID:            user.ID,
		TelegramMessageID: int64(c.Message().ID),
		FilePath:          filePath,
		OCRText:           parsedReceipt.RawText,
		MerchantName:      parsedReceipt.MerchantName,
		PurchaseAt:        parsedReceipt.PurchaseAt,
		TotalAmount:       parsedReceipt.TotalAmount,
		Currency:          "RUB",
	}

	err = h.Repo.SaveReceipt(receipt)
	if err != nil {
		log.Println("Error saving receipt:", err)
		return c.Send("Ошибка сохранения чека")
	}

	// Calculate categories
	categoryTotals := make(map[string]float64)

	var receiptItemsToSave []*models.ReceiptItem
	var expensesToSave []*models.Expense

	// Caching categories to avoid DB queries inside the loop
	categoryCache := make(map[string]int)
	getCategoryID := func(catName string) int {
		if id, exists := categoryCache[catName]; exists {
			return id
		}
		cat, err := h.Repo.GetCategoryByName(catName)
		var catID int
		if err == nil {
			catID = cat.ID
		} else {
			cOther, errOther := h.Repo.GetCategoryByName("other")
			if errOther == nil {
				catID = cOther.ID
			}
		}
		categoryCache[catName] = catID
		return catID
	}

	for _, item := range parsedReceipt.Items {
		catName := parser.DetectCategory(item.Name)
		catID := getCategoryID(catName)

		receiptItem := &models.ReceiptItem{
			ReceiptID:  receipt.ID,
			ItemName:   item.Name,
			Amount:     item.Amount,
			CategoryID: catID,
		}
		receiptItemsToSave = append(receiptItemsToSave, receiptItem)

		categoryTotals[catName] += item.Amount

		// Add expense
		exp := &models.Expense{
			UserID:      user.ID,
			CategoryID:  catID,
			SourceType:  "receipt",
			Description: item.Name,
			Amount:      item.Amount,
			ExpenseAt:   parsedReceipt.PurchaseAt,
		}
		expensesToSave = append(expensesToSave, exp)
	}

	if len(receiptItemsToSave) > 0 {
		err = h.Repo.SaveReceiptItems(receiptItemsToSave)
		if err != nil {
			log.Println("Error bulk saving receipt items:", err)
		}
	}

	if len(expensesToSave) > 0 {
		err = h.Repo.AddExpenses(expensesToSave)
		if err != nil {
			log.Println("Error bulk saving expenses:", err)
		}
	}

	// If no items were parsed but there is a total, put it to other
	if len(parsedReceipt.Items) == 0 && parsedReceipt.TotalAmount > 0 {
		cat, _ := h.Repo.GetCategoryByName("other")
		var catID int
		if cat != nil {
			catID = cat.ID
		}

		exp := &models.Expense{
			UserID:      user.ID,
			CategoryID:  catID,
			SourceType:  "receipt",
			Description: "Receipt Total",
			Amount:      parsedReceipt.TotalAmount,
			ExpenseAt:   parsedReceipt.PurchaseAt,
		}
		h.Repo.AddExpense(exp)
		categoryTotals["other"] += parsedReceipt.TotalAmount
	}

	response := fmt.Sprintf("Чек обработан:\nМагазин: %s\nСумма: %.2f RUB\nКатегории:\n", receipt.MerchantName, receipt.TotalAmount)
	for cat, total := range categoryTotals {
		response += fmt.Sprintf("* %s: %.2f\n", parser.GetCategoryInRussian(cat), total)
	}

	return c.Send(response)
}

func (h *BotHandler) HandleReportDay(c telebot.Context) error {
	return h.handleReport(c, "day")
}

func (h *BotHandler) HandleReportWeek(c telebot.Context) error {
	return h.handleReport(c, "week")
}

func (h *BotHandler) HandleReportMonth(c telebot.Context) error {
	return h.handleReport(c, "month")
}

func (h *BotHandler) handleReport(c telebot.Context, period string) error {
	user, err := h.Repo.GetOrCreateUser(c.Sender().ID, c.Sender().Username)
	if err != nil {
		return c.Send("Ошибка пользователя")
	}

	report, err := h.Repo.GetExpensesReport(user.ID, period)
	if err != nil {
		return c.Send("Ошибка генерации отчета")
	}

	if len(report) == 0 {
		return c.Send(fmt.Sprintf("Нет трат за период: %s", period))
	}

	res := fmt.Sprintf("Траты за %s:\n", period)
	var total float64
	for _, row := range report {
		res += fmt.Sprintf("%s: %.2f\n", parser.GetCategoryInRussian(row.CategoryName), row.TotalAmount)
		total += row.TotalAmount
	}
	res += fmt.Sprintf("Итого: %.2f", total)

	return c.Send(res)
}

func (h *BotHandler) HandleReportCategory(c telebot.Context) error {
	args := c.Args()
	if len(args) < 2 {
		return c.Send("Использование: /category <категория> <период (month, week, day)>")
	}
	catName := args[0]
	period := args[1]

	user, _ := h.Repo.GetOrCreateUser(c.Sender().ID, c.Sender().Username)
	total, err := h.Repo.GetCategoryExpensesReport(user.ID, catName, period)
	if err != nil {
		return c.Send("Ошибка")
	}

	return c.Send(fmt.Sprintf("Категория: %s\nПериод: %s\nСумма: %.2f", parser.GetCategoryInRussian(catName), period, total))
}

func (h *BotHandler) HandleReportUser(c telebot.Context) error {
	args := c.Args()
	period := "month"
	if len(args) > 0 {
		period = args[0]
	}

	user, _ := h.Repo.GetOrCreateUser(c.Sender().ID, c.Sender().Username)
	total, err := h.Repo.GetUserExpensesReport(user.ID, period)
	if err != nil {
		return c.Send("Ошибка")
	}

	return c.Send(fmt.Sprintf("Пользователь: %s\nПериод: %s\nСумма: %.2f", user.DisplayName, period, total))
}

func (h *BotHandler) HandleLast(c telebot.Context) error {
	user, _ := h.Repo.GetOrCreateUser(c.Sender().ID, c.Sender().Username)
	expenses, err := h.Repo.GetLastExpenses(user.ID, 10)
	if err != nil {
		return c.Send("Ошибка")
	}
	if len(expenses) == 0 {
		return c.Send("Нет последних записей")
	}

	var res strings.Builder
	res.WriteString("Последние 10 записей:\n")
	for _, e := range expenses {
		res.WriteString(fmt.Sprintf("- %s: %.2f (%s)\n", e.Description, e.Amount, e.ExpenseAt.Format("02.01 15:04")))
	}
	return c.Send(res.String())
}

func (h *BotHandler) HandleHelp(c telebot.Context) error {
	helpMsg := `
Добавление траты (текст):
просто напишите <описание> <сумма>, например: хлеб 50

Добавление чека (фото):
отправьте фото чека

Отчеты:
/day - за сегодня
/week - за неделю
/month - за месяц
/category <категория_на_англ> <период> - отчет по категории
/byuser <период> - отчет по пользователю
/last - последние 10 записей
/help - справка
`
	return c.Send(helpMsg)
}
