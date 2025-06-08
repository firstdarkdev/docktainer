package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func updateBranch(cloneURL, branch string, silent bool) {
	// Log and Send discord notification
	if !silent {
		logMessage("Received Push Event for branch: %s", branch)
		sendDiscordMessage(branch, "A new build has started for `"+branch+"`", "Build Started", yellow, "")
	}

	// Set up the Working and Output folders
	repoBranchPath := fmt.Sprintf("%s/%s", repoPath, branch)
	htmlBranchPath := fmt.Sprintf("%s/%s", htmlPath, branch)

	// If the output folder does not exist, we clone it
	if _, err := os.Stat(repoBranchPath); os.IsNotExist(err) {
		cmd := exec.Command("git", "clone", "-b", branch, cloneURL, repoBranchPath)
		if err := cmd.Run(); err != nil {
			logMessage("Error cloning branch %s: %v", branch, err)
			if !silent {
				sendDiscordMessage(branch, "A build has failed for `"+branch+"`", "Build Failed", red, err.Error())
			}
			return
		}
	} else {
		// Output folder does exist, so we force clone it
		cmd := exec.Command("git", "-C", repoBranchPath, "fetch", "--all")
		_ = cmd.Run()
		cmd = exec.Command("git", "-C", repoBranchPath, "reset", "--hard", fmt.Sprintf("origin/%s", branch))
		if err := cmd.Run(); err != nil {
			logMessage("Error resetting branch %s: %v", branch, err)

			if !silent {
				sendDiscordMessage(branch, "A build has failed for `"+branch+"`", "Build Failed", red, err.Error())
			}
			return
		}
	}

	// Check if this is a docusaurus build
	isDocusaurus := false
	if _, err := os.Stat(fmt.Sprintf("%s/docusaurus.config.ts", repoBranchPath)); err == nil {
		isDocusaurus = true
	}

	// Set up the command
	var cmd *exec.Cmd
	if isDocusaurus {
		// Install dependencies
		cmd = exec.Command("npm", "install")
		cmd.Dir = repoBranchPath
		_ = cmd.Run()

		// Run the actual build
		cmd = exec.Command("npm", "run", "build")
		cmd.Dir = repoBranchPath
	} else {
		// It's a Retype site. Run that instead
		cmd = exec.Command("retype", "build")
		cmd.Dir = repoBranchPath
	}

	// Set up logging for the command output
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Log the error and output (both stdout and stderr)
		logMessage("Build failed for branch %s: %v", branch, err)
		logMessage("Stdout: %s", out.String())
		logMessage("Stderr: %s", stderr.String())

		if !silent {
			sendDiscordMessage(branch, "A build has failed for `"+branch+"`", "Build Failed", red, err.Error())
		}
		return
	}

	// Remove old HTML folder, and copy the new one
	os.RemoveAll(htmlBranchPath)
	os.MkdirAll(htmlBranchPath, os.ModePerm)

	// Copy the output files
	var copyCmd *exec.Cmd
	if isDocusaurus {
		copyCmd = exec.Command("cp", "-r", fmt.Sprintf("%s/build/.", repoBranchPath), htmlBranchPath)
	} else {
		copyCmd = exec.Command("cp", "-r", fmt.Sprintf("%s/.retype/.", repoBranchPath), htmlBranchPath)
	}
	_ = copyCmd.Run()

	// Send final notification to discord
	logMessage("Successfully built and updated branch: %s", branch)

	if !silent {
		sendDiscordMessage(branch, "A build has completed for `"+branch+"`", "Build Successful", green, "")
	}
}

// The Branch was deleted, so we delete the output folder
func deleteBranch(branch string) {
	htmlBranchPath := fmt.Sprintf("%s/%s", htmlPath, branch)

	if _, err := os.Stat(htmlBranchPath); os.IsNotExist(err) {
		return
	}

	logMessage("Deleting branch: %s", branch)
	os.RemoveAll(htmlBranchPath)
	logMessage("Deleted: %s", htmlBranchPath)
	sendDiscordMessage(branch, "A deployment for `"+branch+"` has been deleted", "Branch Deleted", orange, "")
}

// Fetch all remote branches from the GitHub repo
func getAllBranches(cloneURL string) ([]string, error) {
	cmd := exec.Command("git", "ls-remote", "--heads", cloneURL)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to fetch branches: %v", err)
	}

	// Parse the output to get branch names
	branches := []string{}
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) > 1 {
			branch := parts[1]
			branch = strings.TrimPrefix(branch, "refs/heads/") // Remove refs/heads/ prefix
			branches = append(branches, branch)
		}
	}
	return branches, nil
}

// Check if a branch folder exists in /app/html
func branchFolderExists(branch string) bool {
	htmlBranchPath := fmt.Sprintf("%s/%s", htmlPath, branch)
	if _, err := os.Stat(htmlBranchPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// Initialize branches on first run
func initializeBranches() {
	branches, err := getAllBranches(baseRepository)
	if err != nil {
		logMessage("Error fetching branches: %v", err)
		return
	}

	// For each branch, check if it exists in the HTML folder. If not, update it
	for _, branch := range branches {
		if !branchFolderExists(branch) {
			logMessage("Found missing branch: %s. Building...", branch)
			updateBranch(baseRepository, branch, true)
		}
	}
}
