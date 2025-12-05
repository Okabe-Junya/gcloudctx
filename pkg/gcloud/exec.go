package gcloud

import (
	"fmt"
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
