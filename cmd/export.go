package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Okabe-Junya/gcloudctx/internal/output"
	"github.com/Okabe-Junya/gcloudctx/pkg/gcloud"
	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
)

var (
	exportFormatFlag string
	exportOutputFlag string
)

// ExportConfig represents the exported configuration format.
type ExportConfig struct {
	Name    string `json:"name"              yaml:"name"`
	Account string `json:"account,omitempty" yaml:"account,omitempty"`
	Project string `json:"project,omitempty" yaml:"project,omitempty"`
	Region  string `json:"region,omitempty"  yaml:"region,omitempty"`
	Zone    string `json:"zone,omitempty"    yaml:"zone,omitempty"`
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
		currentConfig, err := gcloud.GetActiveConfiguration()
		if err != nil {
			output.PrintError(err.Error(), !noColorFlag)

			return err
		}

		configName = currentConfig.Name
	} else {
		configName = args[0]
	}

	config, err := gcloud.GetConfigurationInfo(configName)
	if err != nil {
		output.PrintError(err.Error(), !noColorFlag)

		return err
	}

	exportConfig := ExportConfig{
		Name:    config.Name,
		Account: config.Properties.Core.Account,
		Project: config.Properties.Core.Project,
		Region:  config.Properties.Compute.Region,
		Zone:    config.Properties.Compute.Zone,
	}

	data, err := marshalExport(exportConfig, exportFormatFlag)
	if err != nil {
		output.PrintError(err.Error(), !noColorFlag)

		return err
	}

	if exportOutputFlag != "" {
		if err := os.WriteFile(exportOutputFlag, data, 0o600); err != nil {
			output.PrintError(fmt.Sprintf("failed to write file: %v", err), !noColorFlag)

			return fmt.Errorf("failed to write file: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("exported configuration %q to %s", configName, exportOutputFlag), !noColorFlag)
	} else {
		fmt.Print(string(data))
	}

	return nil
}

func marshalExport(cfg ExportConfig, format string) ([]byte, error) {
	switch format {
	case "yaml", "yml":
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal configuration: %w", err)
		}

		return data, nil
	case "json":
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal configuration: %w", err)
		}

		return append(data, '\n'), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s (use yaml or json)", format)
	}
}
