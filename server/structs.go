package main

import "github.com/disgoorg/disgo/webhook"

// GitHubWebHook GitHub webhook data
type GitHubWebHook struct {
	Ref  string `json:"ref"`
	Repo struct {
		CloneURL string `json:"clone_url"`
	} `json:"repository"`
	Deleted bool `json:"deleted"`
}

// Working directories
const (
	repoPath = "/app/repos"
	htmlPath = "/app/html"
	sslPath  = "/app/ssl"
	logFile  = "/app/webhook.log"
)

// Env variables
var (
	webhookSecret     string
	discordWebhookUrl string
	baseUrl           string
	discordClient     webhook.Client
)

// Embed Colors
const (
	yellow = 0xFFFF00 // Yellow
	green  = 0x00FF00 // Green
	red    = 0xFF0000 // Red
	orange = 0xFFA500
)
