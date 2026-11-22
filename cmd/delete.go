package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Okabe-Junya/gcloudctx/internal/output"
	"github.com/Okabe-Junya/gcloudctx/pkg/gcloud"
	"github.com/spf13/cobra"
)

var (
	forceFlag bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete <configuration-name>",
	Short: "Delete a gcloud configuration",
	Long: `Delete a gcloud configuration.

You cannot delete the currently active configuration.
Use -f/--force to skip the confirmation prompt.

Examples:
  gcloudctx delete my-old-config
  gcloudctx delete my-old-config --force`,
	Args:              cobra.ExactArgs(1),
	RunE:              runDelete,
	ValidArgsFunction: completeConfigNamesForDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Skip confirmation prompt")
	rootCmd.AddCommand(deleteCmd)
}

func completeConfigNamesForDelete(
	cmd *cobra.Command, args []string, toComplete string,
) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	configs, err := gcloud.ListConfigurations()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var names []string

	for _, config := range configs {
		// The active configuration cannot be deleted, so do not suggest it.
		if !config.IsActive {
			names = append(names, config.Name)
		}
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

func runDelete(cmd *cobra.Command, args []string) error {
	configName := args[0]

	// No explicit CheckGcloudInstalled: every pkg/gcloud call runs it first,
	// so the prompt can appear before a missing gcloud is reported.
	if !forceFlag {
		fmt.Printf("Are you sure you want to delete configuration %q? (y/N): ", configName)

		reader := bufio.NewReader(os.Stdin)

		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Deletion canceled")

			return nil
		}
	}

	if err := gcloud.DeleteConfiguration(configName); err != nil {
		output.PrintError(err.Error(), !noColorFlag)

		return err
	}

	output.PrintSuccess(fmt.Sprintf("deleted configuration %q", configName), !noColorFlag)

	return nil
}
