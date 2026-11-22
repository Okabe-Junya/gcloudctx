package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Okabe-Junya/gcloudctx/internal/output"
	"github.com/Okabe-Junya/gcloudctx/pkg/gcloud"
	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
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

// parseImportData decodes an exported configuration from file contents, using
// ext to pick the decoder and falling back to content sniffing when unknown.
func parseImportData(data []byte, ext string) (ExportConfig, error) {
	// Strip a UTF-8 BOM: neither goccy/go-yaml nor encoding/json skips it, so a
	// leading BOM would otherwise be folded into the first mapping key and that
	// field would silently decode as empty.
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	var (
		cfg ExportConfig
		err error
	)

	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &cfg)
	case ".json":
		err = json.Unmarshal(data, &cfg)
	default:
		if err = yaml.Unmarshal(data, &cfg); err != nil {
			err = json.Unmarshal(data, &cfg)
		}
	}

	if err != nil {
		return ExportConfig{}, fmt.Errorf("failed to parse file: %w", err)
	}

	return cfg, nil
}

func readImportFile(filePath string) (ExportConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		output.PrintError(fmt.Sprintf("failed to read file: %v", err), !noColorFlag)

		return ExportConfig{}, fmt.Errorf("failed to read file: %w", err)
	}

	importConfig, err := parseImportData(data, strings.ToLower(filepath.Ext(filePath)))
	if err != nil {
		output.PrintError(err.Error(), !noColorFlag)

		return ExportConfig{}, err
	}

	return importConfig, nil
}

func resolveImportName(importConfig ExportConfig) (string, error) {
	configName := importConfig.Name
	if importNameFlag != "" {
		configName = importNameFlag
	}

	if configName == "" {
		output.PrintError("configuration name is required (use --name or include 'name' in the file)", !noColorFlag)

		return "", errors.New("missing configuration name")
	}

	if err := gcloud.ValidateConfigurationName(configName); err != nil {
		output.PrintError(err.Error(), !noColorFlag)

		return "", err
	}

	return configName, nil
}

func writeImportedConfiguration(configName string, importConfig ExportConfig) error {
	if gcloud.ConfigurationExists(configName) {
		if !importOverwriteFlag {
			output.PrintError(
				fmt.Sprintf("configuration %q already exists (use --overwrite to replace)", configName), !noColorFlag)

			return errors.New("configuration already exists")
		}

		// DeleteConfiguration refuses the active configuration, so --overwrite
		// fails when the target is the active one.
		if err := gcloud.DeleteConfiguration(configName); err != nil {
			output.PrintError(fmt.Sprintf("failed to delete existing configuration: %v", err), !noColorFlag)

			return err
		}
	}

	if err := gcloud.CreateConfiguration(configName); err != nil {
		output.PrintError(err.Error(), !noColorFlag)

		return err
	}

	err := gcloud.SetConfigProperties(
		configName, importConfig.Account, importConfig.Project, importConfig.Region, importConfig.Zone)
	if err != nil {
		// Report the property error as the cause; a cleanup failure is only a warning.
		if cleanupErr := gcloud.DeleteConfiguration(configName); cleanupErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to cleanup configuration: %v\n", cleanupErr)
		}

		output.PrintError(err.Error(), !noColorFlag)

		return err
	}

	return nil
}

func runImport(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	importConfig, err := readImportFile(filePath)
	if err != nil {
		return err
	}

	configName, err := resolveImportName(importConfig)
	if err != nil {
		return err
	}

	if err := writeImportedConfiguration(configName, importConfig); err != nil {
		return err
	}

	output.PrintSuccess(fmt.Sprintf("imported configuration %q from %s", configName, filePath), !noColorFlag)

	if importActivateFlag {
		if err := gcloud.ActivateConfiguration(configName); err != nil {
			output.PrintError(err.Error(), !noColorFlag)

			return err
		}

		output.PrintSuccess(fmt.Sprintf("activated configuration %q", configName), !noColorFlag)
	}

	return nil
}
