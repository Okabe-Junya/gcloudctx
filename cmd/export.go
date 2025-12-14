package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Okabe-Junya/gcloudctx/internal/output"
	"github.com/Okabe-Junya/gcloudctx/pkg/gcloud"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	exportFormatFlag string
	exportOutputFlag string
)

// ExportConfig represents the exported configuration format
type ExportConfig struct {
	Name    string `json:"name" yaml:"name"`
	Account string `json:"account,omitempty" yaml:"account,omitempty"`
	Project string `json:"project,omitempty" yaml:"project,omitempty"`
	Region  string `json:"region,omitempty" yaml:"region,omitempty"`
	Zone    string `json:"zone,omitempty" yaml:"zone,omitempty"`
}

var exportCmd = &cobra.Command{
	Use:   "export [configuration-name]",
	Short: "Export a gcloud configuration to a file",
	Long: `Export a gcloud configuration to YAML or JSON format.

The exported file can be used to import the configuration on another machine
or share it with team members.

Examples:
  gcloudctx export production                    # Export to stdout (YAML)
  gcloudctx export production -o config.yaml     # Export to file
  gcloudctx export production --format json      # Export as JSON
  gcloudctx export                               # Export current configuration`,
	Args:              cobra.MaximumNArgs(1),
	RunE:              runExport,
	ValidArgsFunction: completeConfigNames,
}

func init() {
	exportCmd.Flags().StringVarP(&exportFormatFlag, "format", "f", "yaml", "Output format (yaml or json)")
	exportCmd.Flags().StringVarP(&exportOutputFlag, "output", "o", "", "Output file (defaults to stdout)")
	rootCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	var configName string

	if len(args) == 0 {
		// Export current configuration
		currentConfig, err := gcloud.GetActiveConfiguration()
		if err != nil {
			output.PrintError(err.Error(), !noColorFlag)
			return err
		}
		configName = currentConfig.Name
	} else {
		configName = args[0]
	}

	// Get configuration info
	config, err := gcloud.GetConfigurationInfo(configName)
	if err != nil {
		output.PrintError(err.Error(), !noColorFlag)
		return err
	}

	// Build export structure
	exportConfig := ExportConfig{
		Name:    config.Name,
		Account: config.Properties.Core.Account,
		Project: config.Properties.Core.Project,
		Region:  config.Properties.Compute.Region,
		Zone:    config.Properties.Compute.Zone,
	}

	// Marshal to the requested format
	var data []byte
	switch exportFormatFlag {
	case "yaml", "yml":
		data, err = yaml.Marshal(exportConfig)
	case "json":
		data, err = json.MarshalIndent(exportConfig, "", "  ")
		if err == nil {
			data = append(data, '\n')
		}
	default:
		output.PrintError(fmt.Sprintf("unsupported format: %s (use yaml or json)", exportFormatFlag), !noColorFlag)
		return fmt.Errorf("unsupported format")
	}

	if err != nil {
		output.PrintError(fmt.Sprintf("failed to marshal configuration: %v", err), !noColorFlag)
		return err
	}

	// Write output
	if exportOutputFlag != "" {
		if err := os.WriteFile(exportOutputFlag, data, 0o644); err != nil {
			output.PrintError(fmt.Sprintf("failed to write file: %v", err), !noColorFlag)
			return err
		}
		output.PrintSuccess(fmt.Sprintf("exported configuration %q to %s", configName, exportOutputFlag), !noColorFlag)
	} else {
		fmt.Print(string(data))
	}

	return nil
}
