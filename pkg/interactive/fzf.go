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

func getSelfCommand() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(executable)
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

	// Lines use the "* name (account) [project]" shape ParseConfigurationName expects.
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

	selfCmd, err := getSelfCommand()
	if err != nil {
		selfCmd = "gcloudctx"
	}

	fzfArgs := buildFzfArgs(selfCmd)
	cmd := exec.Command("fzf", fzfArgs...)

	cmd.Stdin = strings.NewReader(inputBuilder.String())
	cmd.Stderr = os.Stderr

	var output bytes.Buffer
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		// fzf exits with 130 when the user cancels with ESC or Ctrl-C.
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 130 {
				return "", ErrSelectionCanceled
			}
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
