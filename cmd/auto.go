package cmd

import (
	"errors"
	"fmt"
	"os"

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
	configName, dir, err := local.FindLocalConfig()
	if err != nil {
		// Print to stderr so shell cd hooks (which redirect 2>/dev/null) stay silent,
		// while manual invocations show feedback.
		fmt.Fprintln(os.Stderr, "no .gcloudctx file found in current or parent directories")

		return nil //nolint:nilerr // a directory without .gcloudctx is the normal case for cd hooks
	}

	if !gcloud.ConfigurationExists(configName) {
		output.PrintError(fmt.Sprintf("configuration %q (from %s/.gcloudctx) does not exist", configName, dir), !noColorFlag)

		return errors.New("configuration not found")
	}

	currentConfig, err := gcloud.GetActiveConfiguration()
	if err != nil {
		output.PrintError(err.Error(), !noColorFlag)

		return err
	}

	if currentConfig.Name == configName {
		return nil
	}

	if err := history.SavePreviousConfig(currentConfig.Name); err != nil {
		fmt.Printf("Warning: failed to save history: %v\n", err)
	}

	if err := gcloud.ActivateConfiguration(configName); err != nil {
		output.PrintError(err.Error(), !noColorFlag)

		return err
	}

	output.PrintSuccess(fmt.Sprintf("switched to configuration %q (from %s)", configName, dir), !noColorFlag)

	return nil
}
