// Package interactive provides interactive selection functionality using fzf.
// It enables users to browse and select gcloud configurations with a fuzzy finder
// interface, including live preview of configuration details.
package interactive

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Okabe-Junya/gcloudctx/pkg/gcloud"
)

// IsFzfInstalled checks if fzf is installed
func IsFzfInstalled() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}

// getSelfCommand returns the path to the current executable
func getSelfCommand() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	// Resolve symlinks
	return filepath.EvalSymlinks(executable)
}

// SelectConfigurationInteractive allows the user to select a configuration using fzf
// This implementation passes data via stdin and uses Go for preview (no shell commands)
func SelectConfigurationInteractive(configs []gcloud.Configuration, currentConfig string) (string, error) {
	if !IsFzfInstalled() {
		return "", ErrFzfNotInstalled
	}

	if len(configs) == 0 {
		return "", ErrNoConfigurations
	}

	// Build the input data for fzf (format: "* name (account) [project]")
	var inputBuilder strings.Builder
	for _, config := range configs {
		marker := " "
		if config.Name == currentConfig {
			marker = "*"
		}

		line := fmt.Sprintf("%s %s", marker, config.Name)
		if config.Properties.Core.Account != "" {
			line += fmt.Sprintf(" (%s)", config.Properties.Core.Account)
		}
		if config.Properties.Core.Project != "" {
			line += fmt.Sprintf(" [%s]", config.Properties.Core.Project)
		}

		inputBuilder.WriteString(line + "\n")
	}

	// Get the path to the current executable for preview
	selfCmd, err := getSelfCommand()
	if err != nil {
		// Fallback to "gcloudctx" if we can't get the executable path
		selfCmd = "gcloudctx"
	}

	// Build fzf command arguments (preview uses Go command, no shell!)
	fzfArgs := buildFzfArgs(selfCmd)
	cmd := exec.Command("fzf", fzfArgs...)

	// Pass data via stdin (no FZF_DEFAULT_COMMAND needed)
	cmd.Stdin = strings.NewReader(inputBuilder.String())
	cmd.Stderr = os.Stderr

	var output bytes.Buffer
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		// User canceled (ESC or Ctrl+C)
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 130 {
				return "", ErrSelectionCanceled
			}
		}
		return "", fmt.Errorf("fzf selection failed: %w", err)
	}

	// Parse the selected line to extract the configuration name
	selected := strings.TrimSpace(output.String())
	if selected == "" {
		return "", ErrNoSelection
	}

	// Extract the configuration name from the formatted line
	return ParseConfigurationName(selected)
}

// buildFzfArgs builds the fzf command arguments
// Preview is handled by a Go command (no shell scripts!)
func buildFzfArgs(selfCmd string) []string {
	// Get custom fzf options from environment
	customOpts := os.Getenv(EnvFzfOptions)

	// Default options
	args := []string{
		"--ansi",
		"--height", getEnvOrDefault(EnvFzfHeight, DefaultFzfHeight),
		"--reverse",
		"--border",
		"--header", "Select a configuration:",
		"--prompt", "gcloud> ",
	}

	// Add preview unless disabled
	if os.Getenv(EnvDisablePreview) != "1" {
		// Use Go command for preview (100% Go, no shell commands at all!)
		// Pass the entire fzf selection line to our preview command
		// It will parse the configuration name internally
		previewCmd := fmt.Sprintf(`%s %s {}`, selfCmd, PreviewCommand)
		args = append(args,
			"--preview", previewCmd,
			"--preview-window", getEnvOrDefault(EnvFzfPreviewWindow, DefaultFzfPreviewWindow),
		)
	}

	// Add custom options if provided
	if customOpts != "" {
		customArgs := strings.Fields(customOpts)
		args = append(args, customArgs...)
	}

	return args
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
