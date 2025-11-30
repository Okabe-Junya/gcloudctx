package cmd

import (
	"fmt"

	"github.com/Okabe-Junya/gcloudctx/internal/output"
	"github.com/Okabe-Junya/gcloudctx/pkg/gcloud"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename <old-name> <new-name>",
	Short: "Rename a gcloud configuration",
	Long: `Rename a gcloud configuration.

This creates a new configuration with the new name, copies all properties
from the old configuration, and deletes the old one.

Examples:
  gcloudctx rename old-config new-config`,
	Args:              cobra.ExactArgs(2),
	RunE:              runRename,
	ValidArgsFunction: completeConfigNamesForRename,
}

func init() {
	rootCmd.AddCommand(renameCmd)
}

// completeConfigNamesForRename provides completion for rename command
func completeConfigNamesForRename(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Only complete the first argument (old name)
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

func runRename(cmd *cobra.Command, args []string) error {
	oldName := args[0]
	newName := args[1]

	// Check if gcloud is installed
	if err := gcloud.CheckGcloudInstalled(); err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	// Rename the configuration
	if err := gcloud.RenameConfiguration(oldName, newName); err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	output.PrintSuccess(fmt.Sprintf("renamed configuration %q to %q", oldName, newName), !noColorFlag)
	return nil
}
