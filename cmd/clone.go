package cmd

import (
	"fmt"

	"github.com/Okabe-Junya/gcloudctx/internal/output"
	"github.com/Okabe-Junya/gcloudctx/pkg/gcloud"
	"github.com/spf13/cobra"
)

var (
	cloneActivateFlag bool
)

var cloneCmd = &cobra.Command{
	Use:   "clone <source-name> <target-name>",
	Short: "Clone an existing gcloud configuration",
	Long: `Clone an existing gcloud configuration to create a new one.

This creates a new configuration with all properties copied from the source.
The source configuration remains unchanged.

Examples:
  gcloudctx clone production production-test
  gcloudctx clone my-config my-config-backup --activate`,
	Args:              cobra.ExactArgs(2),
	RunE:              runClone,
	ValidArgsFunction: completeConfigNamesForClone,
}

func init() {
	cloneCmd.Flags().BoolVar(&cloneActivateFlag, "activate", false, "Activate the newly cloned configuration")
	rootCmd.AddCommand(cloneCmd)
}

// completeConfigNamesForClone provides completion for clone command
func completeConfigNamesForClone(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Only complete the first argument (source name)
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

func runClone(cmd *cobra.Command, args []string) error {
	sourceName := args[0]
	targetName := args[1]

	// Validate target configuration name before making gcloud calls
	if err := gcloud.ValidateConfigurationName(targetName); err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	// Clone the configuration
	if err := gcloud.CloneConfiguration(sourceName, targetName); err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	output.PrintSuccess(fmt.Sprintf("cloned configuration %q to %q", sourceName, targetName), !noColorFlag)

	// Activate if requested
	if cloneActivateFlag {
		if err := gcloud.ActivateConfiguration(targetName); err != nil {
			output.PrintError(err.Error(), !noColorFlag)
			return err
		}
		output.PrintSuccess(fmt.Sprintf("activated configuration %q", targetName), !noColorFlag)
	}

	return nil
}

