// Package interactive provides interactive selection functionality using fzf.
// It enables users to browse and select gcloud configurations with a fuzzy finder
// interface, including live preview of configuration details.
package interactive

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Okabe-Junya/gcloudctx/pkg/gcloud"
)

// IsFzfInstalled checks if fzf is installed.
func IsFzfInstalled() bool {
	_, err := exec.LookPath("fzf")

	return err == nil
}

// fzfExitInterrupted is fzf's exit status when the user cancels with ESC or Ctrl-C.
const fzfExitInterrupted = 130

func getSelfCommand() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to locate executable: %w", err)
	}

	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return "", fmt.Errorf("failed to resolve executable path: %w", err)
	}

	return resolved, nil
}

// formatFzfInput renders one "* name (account) [project]" line per
// configuration, the shape ParseConfigurationName expects back.
func formatFzfInput(configs []gcloud.Configuration, currentConfig string) string {
	var builder strings.Builder

	for _, config := range configs {
		marker := " "
		if config.Name == currentConfig {
			marker = "*"
		}

		builder.WriteString(marker + " " + config.Name)

		if config.Properties.Core.Account != "" {
			builder.WriteString(" (" + config.Properties.Core.Account + ")")
		}

		if config.Properties.Core.Project != "" {
			builder.WriteString(" [" + config.Properties.Core.Project + "]")
		}

		builder.WriteString("\n")
	}

	return builder.String()
}

// SelectConfigurationInteractive lets the user pick a configuration with fzf
// and returns the selected configuration name.
func SelectConfigurationInteractive(configs []gcloud.Configuration, currentConfig string) (string, error) {
	if !IsFzfInstalled() {
		return "", ErrFzfNotInstalled
	}

	if len(configs) == 0 {
		return "", ErrNoConfigurations
	}

	selfCmd, err := getSelfCommand()
	if err != nil {
		selfCmd = "gcloudctx"
	}

	cmd := exec.Command("fzf", buildFzfArgs(selfCmd)...)
	// Feed the list on stdin rather than via FZF_DEFAULT_COMMAND, which fzf runs
	// through a shell, so configuration names and accounts are never
	// interpolated into a shell command.
	cmd.Stdin = strings.NewReader(formatFzfInput(configs, currentConfig))
	cmd.Stderr = os.Stderr

	var output bytes.Buffer

	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == fzfExitInterrupted {
			return "", ErrSelectionCanceled
		}

		return "", fmt.Errorf("fzf selection failed: %w", err)
	}

	selected := strings.TrimSpace(output.String())
	if selected == "" {
		return "", ErrNoSelection
	}

	return ParseConfigurationName(selected)
}

func buildFzfArgs(selfCmd string) []string {
	customOpts := os.Getenv(EnvFzfOptions)

	args := []string{
		"--ansi",
		"--height", getEnvOrDefault(EnvFzfHeight, DefaultFzfHeight),
		"--reverse",
		"--border",
		"--header", "Select a configuration:",
		"--prompt", "gcloud> ",
	}

	if os.Getenv(EnvDisablePreview) != "1" {
		// The preview re-invokes this binary with the whole selection line,
		// which the preview command parses back into a configuration name.
		previewCmd := fmt.Sprintf(`%s %s {}`, selfCmd, PreviewCommand)
		args = append(args,
			"--preview", previewCmd,
			"--preview-window", getEnvOrDefault(EnvFzfPreviewWindow, DefaultFzfPreviewWindow),
		)
	}

	if customOpts != "" {
		customArgs := strings.Fields(customOpts)
		args = append(args, customArgs...)
	}

	return args
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
