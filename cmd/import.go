package cmd

import (
	"bytes"
	"encoding/json"
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

	var cfg ExportConfig
	var err error
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
	return cfg, err
}

func runImport(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	data, err := os.ReadFile(filePath)
	if err != nil {
		output.PrintError(fmt.Sprintf("failed to read file: %v", err), !noColorFlag)
		return err
	}

	importConfig, err := parseImportData(data, strings.ToLower(filepath.Ext(filePath)))
	if err != nil {
		output.PrintError(fmt.Sprintf("failed to parse file: %v", err), !noColorFlag)
		return err
	}

	configName := importConfig.Name
	if importNameFlag != "" {
		configName = importNameFlag
	}

	if configName == "" {
		output.PrintError("configuration name is required (use --name or include 'name' in the file)", !noColorFlag)
		return fmt.Errorf("missing configuration name")
	}

	if err := gcloud.ValidateConfigurationName(configName); err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	if gcloud.ConfigurationExists(configName) {
		if !importOverwriteFlag {
			output.PrintError(fmt.Sprintf("configuration %q already exists (use --overwrite to replace)", configName), !noColorFlag)
			return fmt.Errorf("configuration already exists")
		}
		if err := gcloud.DeleteConfiguration(configName); err != nil {
			output.PrintError(fmt.Sprintf("failed to delete existing configuration: %v", err), !noColorFlag)
			return err
		}
	}

	if err := gcloud.CreateConfiguration(configName); err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	if err := gcloud.SetConfigProperties(configName, importConfig.Account, importConfig.Project, importConfig.Region, importConfig.Zone); err != nil {
		// Report the property error as the cause; a cleanup failure is only a warning.
		if cleanupErr := gcloud.DeleteConfiguration(configName); cleanupErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to cleanup configuration: %v\n", cleanupErr)
		}
		output.PrintError(err.Error(), !noColorFlag)
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
