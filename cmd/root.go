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
	// Version is the version of the application, set during build via ldflags.
	Version = "dev"
	// Commit is the git commit hash, set during build via ldflags.
	Commit = "none"
	// Date is the build date, set during build via ldflags.
	Date = "unknown"

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
	SilenceErrors:         true,
	SilenceUsage:          true,
}

func init() {
	flags := rootCmd.Flags()
	flags.BoolVarP(&listFlag, "list", "l", false, "List all configurations")
	flags.BoolVarP(&currentFlag, "current", "c", false, "Show current configuration")
	flags.BoolVarP(&interactiveFlag, "interactive", "i", false, "Interactive mode with fzf")
	flags.BoolVar(&syncADCFlag, "sync-adc", false, "Sync Application Default Credentials after switching")
	flags.StringVar(&impersonateFlag, "impersonate-service-account", "", "Service account to impersonate for ADC")
	flags.BoolVar(&showInfoFlag, "info", false, "Show detailed configuration information")
	flags.BoolVar(&noColorFlag, "no-color", false, "Disable colored output")
	flags.StringVarP(&outputFormatFlag, "output", "o", "", "Output format (json, yaml, wide, name)")
}

func runRoot(cmd *cobra.Command, args []string) error {
	if err := gcloud.CheckGcloudInstalled(); err != nil {
		output.PrintError(err.Error(), !noColorFlag)

		return err
	}

	if listFlag {
		return listConfigurations()
	}

	if currentFlag {
		return showCurrentConfiguration()
	}

	if interactiveFlag {
		return interactiveSelection()
	}

	if len(args) == 0 {
		if os.Getenv(interactive.EnvIgnoreFzf) != "1" && interactive.IsFzfInstalled() {
			return interactiveSelection()
		}

		return showCurrentConfiguration()
	}

	targetConfig := args[0]

	if targetConfig == "-" {
		return switchToPrevious()
	}

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

	currentConfig, err := gcloud.GetActiveConfigurationFromList(configs)
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
	// One list call serves both the active-config lookup and the existence
	// check, because every gcloud invocation is slow.
	configs, err := gcloud.ListConfigurations()
	if err != nil {
		output.PrintError(err.Error(), !noColorFlag)

		return err
	}

	currentConfig, err := gcloud.GetActiveConfigurationFromList(configs)
	if err != nil {
		output.PrintError(err.Error(), !noColorFlag)

		return err
	}

	if !gcloud.ConfigurationExistsInList(configs, targetName) {
		output.PrintError(fmt.Sprintf("configuration %q not found", targetName), !noColorFlag)

		return errors.New("configuration not found")
	}

	if currentConfig.Name == targetName {
		output.PrintSuccess(fmt.Sprintf("already on configuration %q", targetName), !noColorFlag)

		return nil
	}

	// History is a convenience for "gcloudctx -"; failing to save it must not
	// block the switch itself.
	if err := history.SavePreviousConfig(currentConfig.Name); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save history: %v\n", err)
	}

	if err := gcloud.ActivateConfiguration(targetName); err != nil {
		output.PrintError(err.Error(), !noColorFlag)

		return err
	}

	output.PrintSuccess(fmt.Sprintf("switched to configuration %q", targetName), !noColorFlag)

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

func completeConfigNames(
	cmd *cobra.Command, args []string, toComplete string,
) ([]string, cobra.ShellCompDirective) {
	// Only the first argument names an existing configuration; a second one
	// (clone/rename target) is new by definition.
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

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

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

const shortCommitLength = 7

func buildVersionString() string {
	result := Version
	if Commit != "none" {
		result += " (commit: " + Commit[:min(shortCommitLength, len(Commit))] + ")"
	}

	if Date != "unknown" {
		result += " built at " + Date
	}

	return result
}
