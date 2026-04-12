package main

import (
	"log"
	"os"

	"expense-bot/internal/bot"
	"expense-bot/internal/config"
	"expense-bot/internal/db"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	repo, err := db.NewRepository(cfg.DBURL)
	if err != nil {
		log.Fatalf("Error connecting to DB: %v", err)
	}
	defer repo.DB.Close()

	botHandler, err := bot.SetupBot(cfg, repo)
	if err != nil {
		log.Fatalf("Error setting up bot: %v", err)
	}

	botHandler.RegisterHandlers()

	log.Printf("Starting bot with webhook on port %s...", cfg.WebhookPort)

	// Create downloads directory just in case it's not created by other mechanisms
	err = os.MkdirAll("downloads", os.ModePerm)
	if err != nil {
		log.Printf("Failed to create downloads directory: %v", err)
	}

	botHandler.Bot.Start()
}
