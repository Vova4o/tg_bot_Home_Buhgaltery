package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken string
	DBURL         string
	AllowedUsers  map[int64]bool
	WebhookURL    string
	WebhookPort   string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load() // ignore error if .env doesn't exist

	allowedUsersStr := os.Getenv("ALLOWED_USERS")
	allowedUsers := make(map[int64]bool)

	if allowedUsersStr != "" {
		for _, u := range strings.Split(allowedUsersStr, ",") {
			id, err := strconv.ParseInt(strings.TrimSpace(u), 10, 64)
			if err == nil {
				allowedUsers[id] = true
			}
		}
	} else {
		log.Println("WARNING: ALLOWED_USERS not set. Bot will accept messages from anyone.")
	}

	return &Config{
		TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
		DBURL:         os.Getenv("DB_URL"),
		AllowedUsers:  allowedUsers,
		WebhookURL:    os.Getenv("WEBHOOK_URL"),
		WebhookPort:   os.Getenv("WEBHOOK_PORT"),
	}, nil
}
