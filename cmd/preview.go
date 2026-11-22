package cmd

import (
	"fmt"

	"github.com/Okabe-Junya/gcloudctx/pkg/gcloud"
	"github.com/Okabe-Junya/gcloudctx/pkg/interactive"
	"github.com/spf13/cobra"
)

var previewCmd = &cobra.Command{
	Use:    interactive.PreviewCommand + " <configuration-name>",
	Short:  "Internal command for fzf preview (do not use directly)",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	RunE:   runPreview,
}

func init() {
	rootCmd.AddCommand(previewCmd)
}

// runPreview reports failures as preview text and returns nil, because its
// only consumer is fzf's preview window, which displays the command's output.
func runPreview(cmd *cobra.Command, args []string) error {
	input := args[0]

	configName, err := interactive.ParseConfigurationName(input)
	if err != nil {
		fmt.Printf("Configuration: %s\n\n(Could not parse configuration name)\n", input)

		return nil //nolint:nilerr // reported as preview text, see above
	}

	config, err := gcloud.GetConfigurationInfo(configName)
	if err != nil {
		fmt.Printf("Configuration: %s\n\n(Details unavailable)\n", configName)

		return nil //nolint:nilerr // reported as preview text, see above
	}

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
