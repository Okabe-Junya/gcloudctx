package cmd

import (
	"fmt"
	"strings"

	"github.com/Okabe-Junya/gcloudctx/pkg/gcloud"
	"github.com/Okabe-Junya/gcloudctx/pkg/interactive"
	"github.com/spf13/cobra"
)

// previewCmd is an internal command used by fzf for preview functionality
var previewCmd = &cobra.Command{
	Use:    interactive.PreviewCommand + " <configuration-name>",
	Short:  "Internal command for fzf preview (do not use directly)",
	Hidden: true, // Hide from help output
	Args:   cobra.ExactArgs(1),
	RunE:   runPreview,
}

func init() {
	rootCmd.AddCommand(previewCmd)
}

func runPreview(cmd *cobra.Command, args []string) error {
	input := args[0]

	// Parse the configuration name from the fzf selection line
	// Format: "* config-name (account) [project]" or "  config-name (account) [project]"
	configName, err := parseConfigName(input)
	if err != nil {
		fmt.Printf("Configuration: %s\n\n(Could not parse configuration name)\n", input)
		return nil
	}

	// Get configuration info
	config, err := gcloud.GetConfigurationInfo(configName)
	if err != nil {
		fmt.Printf("Configuration: %s\n\n(Details unavailable)\n", configName)
		return nil // Don't return error to avoid breaking fzf
	}

	// Display configuration details
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("  Configuration: %s\n", config.Name)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	if config.IsActive {
		fmt.Printf("  Status:  ✓ Active\n")
	} else {
		fmt.Printf("  Status:  Inactive\n")
	}

	if config.Properties.Core.Account != "" {
		fmt.Printf("  Account: %s\n", config.Properties.Core.Account)
	}

	if config.Properties.Core.Project != "" {
		fmt.Printf("  Project: %s\n", config.Properties.Core.Project)
	}

	if config.Properties.Compute.Region != "" {
		fmt.Printf("  Region:  %s\n", config.Properties.Compute.Region)
	}

	if config.Properties.Compute.Zone != "" {
		fmt.Printf("  Zone:    %s\n", config.Properties.Compute.Zone)
	}

	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return nil
}

// parseConfigName extracts the configuration name from a formatted line
func parseConfigName(line string) (string, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", fmt.Errorf("empty line")
	}

	// Split the line into fields
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid format")
	}

	// Find the configuration name (skip marker and parenthesized/bracketed fields)
	for _, part := range parts {
		// Skip marker and fields that start with ( or [
		if part == "*" || strings.HasPrefix(part, "(") || strings.HasPrefix(part, "[") {
			continue
		}
		// This should be the configuration name
		return part, nil
	}

	return "", fmt.Errorf("could not extract configuration name")
}
