// Package history manages the history of previously used gcloud configurations.
// It stores the last active configuration to enable quick switching with the "-" shorthand.
package history

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const historyFileName = ".gcloudctx_previous"

// GetHistoryFilePath returns the path to the history file
func GetHistoryFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, historyFileName), nil
}

// SavePreviousConfig saves the previous configuration name to the history file
func SavePreviousConfig(name string) error {
	path, err := GetHistoryFilePath()
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, []byte(name), 0o600); err != nil {
		return fmt.Errorf("failed to save previous configuration: %w", err)
	}

	return nil
}

// GetPreviousConfig retrieves the previous configuration name from the history file
func GetPreviousConfig() (string, error) {
	path, err := GetHistoryFilePath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no previous configuration found")
		}
		return "", fmt.Errorf("failed to read previous configuration: %w", err)
	}

	name := strings.TrimSpace(string(data))
	if name == "" {
		return "", fmt.Errorf("no previous configuration found")
	}

	return name, nil
}

// ClearHistory removes the history file
func ClearHistory() error {
	path, err := GetHistoryFilePath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear history: %w", err)
	}

	return nil
}
