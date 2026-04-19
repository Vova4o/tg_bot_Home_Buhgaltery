package bot

import (
	"log"

	"expense-bot/internal/config"
	"expense-bot/internal/db"

	"gopkg.in/telebot.v4"
)

type BotHandler struct {
	Bot  *telebot.Bot
	Repo *db.Repository
	Cfg  *config.Config
}

func AuthMiddleware(cfg *config.Config) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			userID := c.Sender().ID
			if !cfg.AllowedUsers[userID] {
				log.Printf("Unauthorized access attempt from user ID: %d", userID)
				return nil // Drop the update silently
			}
			return next(c)
		}
	}
}

func SetupBot(cfg *config.Config, repo *db.Repository) (*BotHandler, error) {
	webhookEndpoint := &telebot.Webhook{
		Listen: ":" + cfg.WebhookPort,
		Endpoint: &telebot.WebhookEndpoint{
			PublicURL: cfg.WebhookURL,
		},
	}

	pref := telebot.Settings{
		Token:  cfg.TelegramToken,
		Poller: webhookEndpoint,
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		return nil, err
	}

	b.Use(AuthMiddleware(cfg))

	handler := &BotHandler{
		Bot:  b,
		Repo: repo,
		Cfg:  cfg,
	}

	return handler, nil
}
