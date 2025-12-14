package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Okabe-Junya/gcloudctx/internal/output"
	"github.com/Okabe-Junya/gcloudctx/pkg/gcloud"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	importActivateFlag  bool
	importOverwriteFlag bool
	importNameFlag      string
)

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import a gcloud configuration from a file",
	Long: `Import a gcloud configuration from a YAML or JSON file.

This creates a new configuration with the properties specified in the file.
The file format is automatically detected from the extension or content.

Examples:
  gcloudctx import config.yaml                # Import from YAML file
  gcloudctx import config.json                # Import from JSON file
  gcloudctx import config.yaml --activate     # Import and activate
  gcloudctx import config.yaml --name myconf  # Import with a different name
  gcloudctx import config.yaml --overwrite    # Overwrite if exists`,
	Args: cobra.ExactArgs(1),
	RunE: runImport,
}

func init() {
	importCmd.Flags().BoolVar(&importActivateFlag, "activate", false, "Activate the imported configuration")
	importCmd.Flags().BoolVar(&importOverwriteFlag, "overwrite", false, "Overwrite if configuration already exists")
	importCmd.Flags().StringVar(&importNameFlag, "name", "", "Use a different name for the imported configuration")
	rootCmd.AddCommand(importCmd)
}

func runImport(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		output.PrintError(fmt.Sprintf("failed to read file: %v", err), !noColorFlag)
		return err
	}

	// Parse configuration
	var importConfig ExportConfig
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &importConfig)
	case ".json":
		err = json.Unmarshal(data, &importConfig)
	default:
		// Try to detect format from content
		if err = yaml.Unmarshal(data, &importConfig); err != nil {
			err = json.Unmarshal(data, &importConfig)
		}
	}

	if err != nil {
		output.PrintError(fmt.Sprintf("failed to parse file: %v", err), !noColorFlag)
		return err
	}

	// Determine configuration name
	configName := importConfig.Name
	if importNameFlag != "" {
		configName = importNameFlag
	}

	if configName == "" {
		output.PrintError("configuration name is required (use --name or include 'name' in the file)", !noColorFlag)
		return fmt.Errorf("missing configuration name")
	}

	// Validate configuration name
	if err := gcloud.ValidateConfigurationName(configName); err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	// Check if configuration already exists
	if gcloud.ConfigurationExists(configName) {
		if !importOverwriteFlag {
			output.PrintError(fmt.Sprintf("configuration %q already exists (use --overwrite to replace)", configName), !noColorFlag)
			return fmt.Errorf("configuration already exists")
		}
		// Delete existing configuration for overwrite
		if err := gcloud.DeleteConfiguration(configName); err != nil {
			// If it's the active config, we can't delete it
			output.PrintError(fmt.Sprintf("failed to delete existing configuration: %v", err), !noColorFlag)
			return err
		}
	}

	// Create the configuration
	if err := gcloud.CreateConfiguration(configName); err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	// Set properties
	if err := setImportedProperties(configName, &importConfig); err != nil {
		// Clean up on failure - ignore error as we're already in error state
		if cleanupErr := gcloud.DeleteConfiguration(configName); cleanupErr != nil {
			// Log cleanup error but continue with original error
			fmt.Fprintf(os.Stderr, "Warning: failed to cleanup configuration: %v\n", cleanupErr)
		}
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	output.PrintSuccess(fmt.Sprintf("imported configuration %q from %s", configName, filePath), !noColorFlag)

	// Activate if requested
	if importActivateFlag {
		if err := gcloud.ActivateConfiguration(configName); err != nil {
			output.PrintError(err.Error(), !noColorFlag)
			return err
		}
		output.PrintSuccess(fmt.Sprintf("activated configuration %q", configName), !noColorFlag)
	}

	return nil
}

func setImportedProperties(configName string, config *ExportConfig) error {
	if config.Account != "" {
		if err := gcloud.RunGcloudCommandQuiet("config", "set", "account", config.Account, "--configuration", configName); err != nil {
			return fmt.Errorf("failed to set account: %w", err)
		}
	}

	if config.Project != "" {
		if err := gcloud.RunGcloudCommandQuiet("config", "set", "project", config.Project, "--configuration", configName); err != nil {
			return fmt.Errorf("failed to set project: %w", err)
		}
	}

	if config.Region != "" {
		if err := gcloud.RunGcloudCommandQuiet("config", "set", "compute/region", config.Region, "--configuration", configName); err != nil {
			return fmt.Errorf("failed to set region: %w", err)
		}
	}

	if config.Zone != "" {
		if err := gcloud.RunGcloudCommandQuiet("config", "set", "compute/zone", config.Zone, "--configuration", configName); err != nil {
			return fmt.Errorf("failed to set zone: %w", err)
		}
	}

	return nil
}
