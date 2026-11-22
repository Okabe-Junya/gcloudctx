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
	ValidArgsFunction: completeConfigNames,
}

func init() {
	rootCmd.AddCommand(renameCmd)
}

func runRename(cmd *cobra.Command, args []string) error {
	oldName := args[0]
	newName := args[1]

	if err := gcloud.ValidateConfigurationName(newName); err != nil {
		output.PrintError(err.Error(), !noColorFlag)

		return err
	}

	if err := gcloud.RenameConfiguration(oldName, newName); err != nil {
		output.PrintError(err.Error(), !noColorFlag)

		return err
	}

	output.PrintSuccess(fmt.Sprintf("renamed configuration %q to %q", oldName, newName), !noColorFlag)

	return nil
}
