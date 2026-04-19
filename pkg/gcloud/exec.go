package gcloud

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CheckGcloudInstalled checks if gcloud CLI is installed
func CheckGcloudInstalled() error {
	_, err := exec.LookPath("gcloud")
	if err != nil {
		return fmt.Errorf("gcloud CLI is not installed or not in PATH")
	}
	return nil
}

// RunGcloudCommand executes a gcloud command with the given arguments
func RunGcloudCommand(args ...string) (string, error) {
	if err := CheckGcloudInstalled(); err != nil {
		return "", err
	}

	cmd := exec.Command("gcloud", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run gcloud command: %w\nOutput: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

// RunGcloudCommandInteractive executes a gcloud command with stdin/stdout/stderr attached
// This is used for commands that require user interaction (e.g., browser-based authentication)
func RunGcloudCommandInteractive(args ...string) error {
	if err := CheckGcloudInstalled(); err != nil {
		return err
	}

	cmd := exec.Command("gcloud", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run gcloud command: %w", err)
	}

	return nil
}

// RunGcloudCommandQuiet executes a gcloud command and suppresses output
// On error, the stderr output is included in the error message for debugging
func RunGcloudCommandQuiet(args ...string) error {
	if err := CheckGcloudInstalled(); err != nil {
		return err
	}

	cmd := exec.Command("gcloud", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Include stderr in error message for better debugging
		if len(output) > 0 {
			return fmt.Errorf("failed to run gcloud command: %w\nOutput: %s", err, strings.TrimSpace(string(output)))
		}
		return fmt.Errorf("failed to run gcloud command: %w", err)
	}

	return nil
}
