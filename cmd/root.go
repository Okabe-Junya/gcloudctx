// Package cmd implements the command-line interface for gcloudctx.
// It provides commands for switching between gcloud configurations,
// managing configurations, and integrating with interactive tools like fzf.
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/Okabe-Junya/gcloudctx/internal/output"
	"github.com/Okabe-Junya/gcloudctx/pkg/gcloud"
	"github.com/Okabe-Junya/gcloudctx/pkg/history"
	"github.com/Okabe-Junya/gcloudctx/pkg/interactive"
	"github.com/spf13/cobra"
)

var (
	// Version is the version of the application, set during build via ldflags
	Version = "dev"
	// Commit is the git commit hash, set during build via ldflags
	Commit = "none"
	// Date is the build date, set during build via ldflags
	Date = "unknown"

	// Flags
	listFlag         bool
	currentFlag      bool
	interactiveFlag  bool
	syncADCFlag      bool
	impersonateFlag  string
	showInfoFlag     bool
	noColorFlag      bool
	outputFormatFlag string
)

var rootCmd = &cobra.Command{
	Use:   "gcloudctx [configuration-name]",
	Short: "Fast way to switch between gcloud configurations",
	Long: `gcloudctx is a tool to quickly switch between gcloud configurations,
inspired by kubectx/kubens.

Examples:
  gcloudctx                    # Show current configuration
  gcloudctx my-config          # Switch to 'my-config'
  gcloudctx -                  # Switch to previous configuration
  gcloudctx -l                 # List all configurations
  gcloudctx -i                 # Interactive selection with fzf
  gcloudctx my-config --sync-adc  # Switch and sync ADC`,
	Version:               buildVersionString(),
	RunE:                  runRoot,
	Args:                  cobra.MaximumNArgs(1),
	ValidArgsFunction:     completeConfigNames,
	DisableFlagsInUseLine: false,
}

func init() {
	rootCmd.Flags().BoolVarP(&listFlag, "list", "l", false, "List all configurations")
	rootCmd.Flags().BoolVarP(&currentFlag, "current", "c", false, "Show current configuration")
	rootCmd.Flags().BoolVarP(&interactiveFlag, "interactive", "i", false, "Interactive mode with fzf")
	rootCmd.Flags().BoolVar(&syncADCFlag, "sync-adc", false, "Sync Application Default Credentials after switching")
	rootCmd.Flags().StringVar(&impersonateFlag, "impersonate-service-account", "", "Service account to impersonate for ADC")
	rootCmd.Flags().BoolVar(&showInfoFlag, "info", false, "Show detailed configuration information")
	rootCmd.Flags().BoolVar(&noColorFlag, "no-color", false, "Disable colored output")
	rootCmd.Flags().StringVarP(&outputFormatFlag, "output", "o", "", "Output format (json, yaml, wide, name)")
}

func runRoot(cmd *cobra.Command, args []string) error {
	// Check if gcloud is installed
	if err := gcloud.CheckGcloudInstalled(); err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	// Handle list flag
	if listFlag {
		return listConfigurations()
	}

	// Handle current flag
	if currentFlag {
		return showCurrentConfiguration()
	}

	// Handle interactive flag
	if interactiveFlag {
		return interactiveSelection()
	}

	// If no arguments, try interactive mode (if fzf is available), otherwise show current configuration
	if len(args) == 0 {
		// Check if we should skip fzf (via environment variable or explicit flag)
		if os.Getenv(interactive.EnvIgnoreFzf) != "1" && interactive.IsFzfInstalled() {
			return interactiveSelection()
		}
		return showCurrentConfiguration()
	}

	// Switch to specified configuration
	targetConfig := args[0]

	// Handle '-' to switch to previous configuration
	if targetConfig == "-" {
		return switchToPrevious()
	}

	// Switch to the target configuration
	return switchConfiguration(targetConfig)
}

func listConfigurations() error {
	configs, err := gcloud.ListConfigurations()
	if err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	if len(configs) == 0 {
		fmt.Println("No configurations found")
		return nil
	}

	// Validate and use output format
	format, err := output.ValidateOutputFormat(outputFormatFlag)
	if err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	return output.PrintConfigurationsWithFormat(configs, format, !noColorFlag)
}

func showCurrentConfiguration() error {
	config, err := gcloud.GetActiveConfiguration()
	if err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	if showInfoFlag {
		output.PrintConfigurationDetails(config, !noColorFlag)
	} else {
		output.PrintCurrentConfiguration(config, !noColorFlag)
	}

	return nil
}

func interactiveSelection() error {
	if !interactive.IsFzfInstalled() {
		output.PrintError("fzf is not installed. Please install fzf for interactive mode.", !noColorFlag)
		return interactive.ErrFzfNotInstalled
	}

	configs, err := gcloud.ListConfigurations()
	if err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	currentConfig, err := gcloud.GetActiveConfiguration()
	if err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	selected, err := interactive.SelectConfigurationInteractive(configs, currentConfig.Name)
	if err != nil {
		if errors.Is(err, interactive.ErrSelectionCanceled) {
			return nil
		}
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	return switchConfiguration(selected)
}

func switchToPrevious() error {
	previousName, err := history.GetPreviousConfig()
	if err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	return switchConfiguration(previousName)
}

func switchConfiguration(targetName string) error {
	// Get current configuration before switching
	currentConfig, err := gcloud.GetActiveConfiguration()
	if err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	// Check if target configuration exists
	if !gcloud.ConfigurationExists(targetName) {
		output.PrintError(fmt.Sprintf("configuration %q not found", targetName), !noColorFlag)
		return fmt.Errorf("configuration not found")
	}

	// Check if already on target configuration
	if currentConfig.Name == targetName {
		output.PrintSuccess(fmt.Sprintf("already on configuration %q", targetName), !noColorFlag)
		return nil
	}

	// Save current configuration to history
	if err := history.SavePreviousConfig(currentConfig.Name); err != nil {
		// Non-fatal error, just warn
		fmt.Fprintf(os.Stderr, "Warning: failed to save history: %v\n", err)
	}

	// Activate the target configuration
	if err := gcloud.ActivateConfiguration(targetName); err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	output.PrintSuccess(fmt.Sprintf("switched to configuration %q", targetName), !noColorFlag)

	// Sync ADC if requested
	if syncADCFlag {
		fmt.Println("Syncing Application Default Credentials...")
		if err := gcloud.SyncADC(impersonateFlag); err != nil {
			output.PrintError(fmt.Sprintf("failed to sync ADC: %v", err), !noColorFlag)
			return err
		}
		output.PrintSuccess("ADC synced successfully", !noColorFlag)
	}

	return nil
}

// completeConfigNames provides completion for configuration names
func completeConfigNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Try to get configuration names
	configs, err := gcloud.ListConfigurations()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var names []string
	for _, config := range configs {
		names = append(names, config.Name)
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// buildVersionString returns a formatted version string including commit and date
func buildVersionString() string {
	result := Version
	if Commit != "none" {
		result += " (commit: " + Commit[:min(7, len(Commit))] + ")"
	}
	if Date != "unknown" {
		result += " built at " + Date
	}
	return result
}
