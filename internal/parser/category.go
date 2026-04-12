package parser

import (
	"strings"
)

var categoryKeywords = map[string][]string{
	"groceries":     {"хлеб", "молоко", "сыр", "мясо", "овощи", "bread", "milk", "cheese", "meat", "vegetables", "продукты"},
	"transport":     {"такси", "метро", "автобус", "бензин", "taxi", "metro", "bus", "fuel"},
	"cigarettes":    {"сигареты", "табак", "cigarettes", "tobacco"},
	"entertainment": {"кино", "бар", "игры", "cinema", "bar", "games", "развлечения"},
	"pharmacy":      {"аптека", "лекарства", "pharmacy", "medicine"},
	"home":          {"дом", "ремонт", "home", "repair"},
	"cafe":          {"кафе", "ресторан", "кофе", "cafe", "restaurant", "coffee"},
	"subscriptions": {"подписка", "интернет", "связь", "subscription", "internet", "mobile"},
	"clothes":       {"одежда", "обувь", "clothes", "shoes"},
}

func DetectCategory(description string) string {
	lowerDesc := strings.ToLower(strings.TrimSpace(description))
	words := strings.Fields(lowerDesc)

	for _, word := range words {
		for category, keywords := range categoryKeywords {
			for _, keyword := range keywords {
				if word == keyword {
					return category
				}
			}
		}
	}

	return "other"
}

func GetCategoryInRussian(category string) string {
	translations := map[string]string{
		"groceries":     "продукты",
		"transport":     "транспорт",
		"entertainment": "развлечения",
		"cigarettes":    "сигареты",
		"pharmacy":      "аптека",
		"home":          "дом",
		"cafe":          "кафе",
		"subscriptions": "подписки",
		"clothes":       "одежда",
		"other":         "другое",
	}
	if translated, ok := translations[category]; ok {
		return translated
	}
	return category
}
