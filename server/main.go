package main

import (
	"github.com/disgoorg/disgo/webhook"
	"log"
	"os"
)

func main() {
	// Environment Variables
	webhookSecret = os.Getenv("WEBHOOK_SECRET")
	discordWebhookUrl = os.Getenv("DISCORD_WEBHOOK_URL")
	baseUrl = os.Getenv("BASE_URL")
	baseRepository = os.Getenv("BASE_REPOSITORY")

	// Check if a Discord webhook was configured, so that we can send discord messages
	if discordWebhookUrl != "" {
		discordClient, _ = webhook.NewWithURL(discordWebhookUrl)
	}

	// Base URL is required
	if baseUrl == "" {
		log.Fatal("BASE_URL environment variable not set")
	}

	if baseRepository != "" {
		initializeBranches()
	}

	// Start the main server
	logMessage("Starting Docktainer...")
	startWebServer()
}
