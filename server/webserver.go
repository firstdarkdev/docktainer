package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func startWebServer() {
	// Configure the HTTP handlers. Both the Webhook, and webserver
	http.HandleFunc("/webhook", handleWebhook)
	http.HandleFunc("/", handleRequest)

	// SSL Cert paths
	certFile := fmt.Sprintf("%s/cert.pem", sslPath)
	keyFile := fmt.Sprintf("%s/key.pem", sslPath)

	// SSL certs were found, so we use HTTPS
	if _, err := os.Stat(certFile); err == nil {
		logMessage("Found SSL Certificate. Starting HTTPS server...")
		err := http.ListenAndServeTLS(":443", certFile, keyFile, nil)
		if err != nil {
			log.Fatalf("Error starting HTTPS server: %v", err)
		} else {
			logMessage("Started HTTPS server on port 443")
		}
	} else {
		// No SSL certs were found, so we use HTTP
		logMessage("No SSL certificates found, starting HTTP server...")
		err := http.ListenAndServe(":80", nil)
		if err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		} else {
			logMessage("Started HTTP server on port 80")
		}
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	host := r.Host

	// Incoming request is likely from GitHub to trigger a build, we skip handling it here
	if host == "webhook."+baseUrl {
		handleWebhook(w, r)
		return
	}

	// Subdomain visited matches the default branch, we redirect to the baseurl
	if host == "main."+baseUrl {
		http.Redirect(w, r, baseUrl, http.StatusTemporaryRedirect)
		return
	}

	// Route the visited subdomain over to the html folder
	subdomain := getSubdomain(host)
	if subdomain == "" {
		serveFile(w, r, fmt.Sprintf("%s/main", htmlPath), r.URL.Path)
	} else {
		serveFile(w, r, fmt.Sprintf("%s/%s", htmlPath, subdomain), r.URL.Path)
	}
}

// Helper method to extract the subdomain from the URL
func getSubdomain(host string) string {
	parts := strings.Split(host, ".")
	if len(parts) > 2 {
		return parts[0]
	}
	return ""
}

func serveFile(w http.ResponseWriter, r *http.Request, basePath string, requestedPath string) {
	// Clean up the URL path and check if it's a directory (ending with /)
	if strings.HasSuffix(requestedPath, "/") {
		// If it's a folder, try to serve index.html
		indexFile := filepath.Join(basePath, requestedPath, "index.html")
		if _, err := os.Stat(indexFile); err == nil {
			// Serve index.html if it exists
			http.ServeFile(w, r, indexFile)
		} else {
			// If index.html doesn't exist, serve 404
			http.NotFound(w, r)
		}
	} else {
		// If it's not a folder, try to serve the requested file (like something.html)
		file := filepath.Join(basePath, requestedPath)
		if _, err := os.Stat(file); err == nil {
			// Serve the specific file if it exists
			http.ServeFile(w, r, file)
		} else {
			// If the file doesn't exist, serve 404
			http.NotFound(w, r)
		}
	}
}

// Handle the incoming GitHub webhook
func handleWebhook(w http.ResponseWriter, req *http.Request) {
	eventType := req.Header.Get("X-GitHub-Event")

	// GitHub ping when first adding the webhook
	if eventType == "ping" || eventType == "created" {
		w.WriteHeader(http.StatusOK)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}

	// Check that the required signature we configured in .env matches the one sent by GitHub
	if !verifySignature(req, body) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	var payload GitHubWebHook
	json.Unmarshal(body, &payload)

	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")

	// Branch was deleted
	if payload.Deleted {
		deleteBranch(branch)
		return
	}

	// Branch was updated
	updateBranch(payload.Repo.CloneURL, branch, false)
}

// Helper function to verify the Secret GitHub sent us
func verifySignature(r *http.Request, body []byte) bool {
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(body)
	expectedMAC := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expectedMAC), []byte(signature))
}
