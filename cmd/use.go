package cmd

import (
	"fmt"

	"github.com/Okabe-Junya/gcloudctx/internal/output"
	"github.com/Okabe-Junya/gcloudctx/pkg/gcloud"
	"github.com/Okabe-Junya/gcloudctx/pkg/local"
	"github.com/spf13/cobra"
)

var (
	useLocalFlag  bool
	useUnsetFlag  bool
	useSwitchFlag bool
)

var useCmd = &cobra.Command{
	Use:   "use [configuration-name]",
	Short: "Set the configuration for the current directory",
	Long: `Set a gcloud configuration to be used in the current directory.

This creates a .gcloudctx file in the current directory that specifies
which configuration should be used. When you run 'gcloudctx use --switch'
or 'gcloudctx auto', it will automatically switch to this configuration.

Examples:
  gcloudctx use my-project          # Set config for current directory
  gcloudctx use my-project --switch # Set and immediately switch
  gcloudctx use --unset             # Remove the .gcloudctx file
  gcloudctx use                     # Show current directory's config`,
	Args:              cobra.MaximumNArgs(1),
	RunE:              runUse,
	ValidArgsFunction: completeConfigNames,
}

func init() {
	useCmd.Flags().BoolVar(&useLocalFlag, "local", true, "Write to the current directory (default)")
	useCmd.Flags().BoolVar(&useUnsetFlag, "unset", false, "Remove the .gcloudctx file from the current directory")
	useCmd.Flags().BoolVar(&useSwitchFlag, "switch", false, "Switch to the configuration after setting it")
	rootCmd.AddCommand(useCmd)
}

func runUse(cmd *cobra.Command, args []string) error {
	// Handle unset flag
	if useUnsetFlag {
		return unsetLocalConfig()
	}

	// If no arguments, show current local config
	if len(args) == 0 {
		return showLocalConfig()
	}

	configName := args[0]

	// Validate configuration name
	if err := gcloud.ValidateConfigurationName(configName); err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	// Check if configuration exists
	if !gcloud.ConfigurationExists(configName) {
		output.PrintError(fmt.Sprintf("configuration %q does not exist", configName), !noColorFlag)
		return fmt.Errorf("configuration not found")
	}

	// Write local config
	if err := local.WriteLocalConfigCurrent(configName); err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	path, _ := local.GetLocalConfigPath()
	output.PrintSuccess(fmt.Sprintf("set local configuration to %q (saved to %s)", configName, path), !noColorFlag)

	// Switch if requested
	if useSwitchFlag {
		return switchConfiguration(configName)
	}

	return nil
}

func showLocalConfig() error {
	configName, dir, err := local.FindLocalConfig()
	if err != nil {
		output.PrintError("no local configuration found in current directory or parent directories", !noColorFlag)
		return err
	}

	fmt.Printf("Local configuration: %s\n", configName)
	fmt.Printf("Found in: %s\n", dir)
	return nil
}

func unsetLocalConfig() error {
	if !local.LocalConfigExists() {
		output.PrintError("no .gcloudctx file in current directory", !noColorFlag)
		return fmt.Errorf("no local config")
	}

	if err := local.RemoveLocalConfigCurrent(); err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	output.PrintSuccess("removed .gcloudctx file from current directory", !noColorFlag)
	return nil
}

