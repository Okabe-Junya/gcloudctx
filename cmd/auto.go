package cmd

import (
	"fmt"

	"github.com/Okabe-Junya/gcloudctx/internal/output"
	"github.com/Okabe-Junya/gcloudctx/pkg/gcloud"
	"github.com/Okabe-Junya/gcloudctx/pkg/history"
	"github.com/Okabe-Junya/gcloudctx/pkg/local"
	"github.com/spf13/cobra"
)

var autoCmd = &cobra.Command{
	Use:   "auto",
	Short: "Automatically switch to the configuration for the current directory",
	Long: `Automatically detect and switch to the configuration specified in .gcloudctx file.

This command searches for a .gcloudctx file starting from the current directory
and walking up to the root. If found, it switches to the specified configuration.

This is useful for automatically switching configurations when changing directories.
You can add this to your shell's cd hook for automatic switching.

Examples:
  gcloudctx auto              # Switch based on .gcloudctx file
  
  # Add to your shell for automatic switching:
  # Bash/Zsh:
  #   cd() { builtin cd "$@" && gcloudctx auto 2>/dev/null; }
  # Fish:
  #   function cd; builtin cd $argv; and gcloudctx auto 2>/dev/null; end`,
	Args: cobra.NoArgs,
	RunE: runAuto,
}

func init() {
	rootCmd.AddCommand(autoCmd)
}

func runAuto(cmd *cobra.Command, args []string) error {
	// Find local config
	configName, dir, err := local.FindLocalConfig()
	if err != nil {
		// Silent fail - this is expected when no .gcloudctx file exists
		return nil
	}

	// Check if configuration exists
	if !gcloud.ConfigurationExists(configName) {
		output.PrintError(fmt.Sprintf("configuration %q (from %s/.gcloudctx) does not exist", configName, dir), !noColorFlag)
		return fmt.Errorf("configuration not found")
	}

	// Get current configuration
	currentConfig, err := gcloud.GetActiveConfiguration()
	if err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	// Already on the target configuration
	if currentConfig.Name == configName {
		return nil
	}

	// Save current configuration to history
	if err := history.SavePreviousConfig(currentConfig.Name); err != nil {
		// Non-fatal error, just warn
		fmt.Printf("Warning: failed to save history: %v\n", err)
	}

	// Activate the target configuration
	if err := gcloud.ActivateConfiguration(configName); err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	output.PrintSuccess(fmt.Sprintf("switched to configuration %q (from %s)", configName, dir), !noColorFlag)
	return nil
}
