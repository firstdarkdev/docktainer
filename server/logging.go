package main

import (
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"log"
	"os"
	"time"
)

func sendDiscordMessage(branch string, message string, state string, color int, errorLog string) {
	// Discord URL was not set up, so the client is null. We do not continue
	if discordClient == nil {
		return
	}

	icon := "⚡️"
	if color == red {
		icon = "💀"
	} else if color == orange {
		icon = "🗑️"
	} else if color == green {
		icon = "🎉"
	}

	// Set up the embed details for the body
	currentTime := time.Now()
	description := fmt.Sprintf("%s Docktainer: %s %s\r\n\r\nBranch: %s\r\nURL: %s", icon, message, icon, branch, fmt.Sprintf("https://%s.%s", branch, baseUrl))

	if errorLog != "" {
		description += fmt.Sprintf("\r\n\r\nBuild Output:\r\n%s", errorLog)
	}

	// Configure the embed
	embed := discord.Embed{
		Title:       fmt.Sprintf("%s %s", branch, state),
		Color:       color,
		Description: description,
		Footer: &discord.EmbedFooter{
			Text: "docktainer",
		},
		Timestamp: &currentTime,
	}

	// Send the embed to discord
	_, err := discordClient.CreateMessage(discord.NewWebhookMessageCreateBuilder().SetContent("").SetEmbeds(embed).Build())

	if err != nil {
		logMessage("Failed to send discord message", err)
	}
}

// Helper function to log to a file, and the console
func logMessage(format string, v ...interface{}) {
	log.Printf(format, v...)
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer file.Close()
		logger := log.New(file, "", log.LstdFlags)
		logger.Printf(format, v...)
	}
}
