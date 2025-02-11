package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func updateBranch(cloneURL, branch string) {
	// Log and Send discord notification
	logMessage("Received Push Event for branch: %s", branch)
	sendDiscordMessage(branch, "A new build has started for "+branch, "Build Started", yellow, "")

	// Set up the Working and Output folders
	repoBranchPath := fmt.Sprintf("%s/%s", repoPath, branch)
	htmlBranchPath := fmt.Sprintf("%s/%s", htmlPath, branch)

	// If the output folder does not exist, we clone it
	if _, err := os.Stat(repoBranchPath); os.IsNotExist(err) {
		cmd := exec.Command("git", "clone", "-b", branch, cloneURL, repoBranchPath)
		if err := cmd.Run(); err != nil {
			logMessage("Error cloning branch %s: %v", branch, err)
			sendDiscordMessage(branch, "A build has failed for "+branch, "Build Failed", red, err.Error())
			return
		}
	} else {
		// Output folder does exist, so we force clone it
		cmd := exec.Command("git", "-C", repoBranchPath, "fetch", "--all")
		_ = cmd.Run()
		cmd = exec.Command("git", "-C", repoBranchPath, "reset", "--hard", fmt.Sprintf("origin/%s", branch))
		if err := cmd.Run(); err != nil {
			logMessage("Error resetting branch %s: %v", branch, err)
			sendDiscordMessage(branch, "A build has failed for "+branch, "Build Failed", red, err.Error())
			return
		}
	}

	// Run the retype build command
	cmd := exec.Command("retype", "build")
	cmd.Dir = repoBranchPath

	// Set up logging for the command output
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Log the error and output (both stdout and stderr)
		logMessage("Retype build failed for branch %s: %v", branch, err)
		logMessage("Stdout: %s", out.String())
		logMessage("Stderr: %s", stderr.String())

		finalLog := out.String()

		// Send failure notification to Discord
		if stderr.String() != "" {
			finalLog = stderr.String()
		}

		sendDiscordMessage(branch, "A build has failed for "+branch, "Build Failed", red, finalLog)
		return
	}

	// Remove old HTML folder, and copy the new one
	os.RemoveAll(htmlBranchPath)
	os.MkdirAll(htmlBranchPath, os.ModePerm)
	exec.Command("cp", "-r", fmt.Sprintf("%s/.retype/.", repoBranchPath), htmlBranchPath).Run()

	// Send final notification to discord
	logMessage("Successfully built and updated branch: %s", branch)
	sendDiscordMessage(branch, "A build has completed for "+branch, "Build Successful", green, "")
}

// Branch was deleted, so we delete the output folder
func deleteBranch(branch string) {
	htmlBranchPath := fmt.Sprintf("%s/%s", htmlPath, branch)

	if _, err := os.Stat(htmlBranchPath); os.IsNotExist(err) {
		return
	}

	logMessage("Deleting branch: %s", branch)
	os.RemoveAll(htmlBranchPath)
	logMessage("Deleted: %s", htmlBranchPath)
	sendDiscordMessage(branch, "A deployment for "+branch+"has been deleted", "Branch Deleted", orange, "")
}
