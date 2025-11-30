package cmd

import (
	"fmt"

	"github.com/Okabe-Junya/gcloudctx/internal/output"
	"github.com/Okabe-Junya/gcloudctx/pkg/gcloud"
	"github.com/spf13/cobra"
)

var (
	activateFlag bool
)

var createCmd = &cobra.Command{
	Use:   "create <configuration-name>",
	Short: "Create a new gcloud configuration",
	Long: `Create a new gcloud configuration.

The new configuration will be created and optionally activated.

Examples:
  gcloudctx create my-new-config
  gcloudctx create my-new-config --activate`,
	Args: cobra.ExactArgs(1),
	RunE: runCreate,
}

func init() {
	createCmd.Flags().BoolVar(&activateFlag, "activate", false, "Activate the newly created configuration")
	rootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	configName := args[0]

	// Check if gcloud is installed
	if err := gcloud.CheckGcloudInstalled(); err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	// Create the configuration
	if err := gcloud.CreateConfiguration(configName); err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	output.PrintSuccess(fmt.Sprintf("created configuration %q", configName), !noColorFlag)

	// Activate if requested
	if activateFlag {
		if err := gcloud.ActivateConfiguration(configName); err != nil {
			output.PrintError(err.Error(), !noColorFlag)
			return err
		}
		output.PrintSuccess(fmt.Sprintf("activated configuration %q", configName), !noColorFlag)
	}

	return nil
}
