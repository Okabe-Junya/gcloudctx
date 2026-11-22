// Package output provides formatting and display utilities for gcloudctx output.
// It handles colorized printing, configuration display, and message formatting.
package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Okabe-Junya/gcloudctx/pkg/gcloud"
	"github.com/fatih/color"
	"github.com/goccy/go-yaml"
)

// Format represents the output format type.
type Format string

// Output format constants.
const (
	FormatDefault Format = ""
	FormatJSON    Format = "json"
	FormatYAML    Format = "yaml"
	FormatWide    Format = "wide"
	FormatName    Format = "name"
)

const ellipsis = "..."

const (
	wideRowFormat    = "%s  %-*s  %-*s  %-*s  %-*s  %s\n"
	wideNameWidth    = 20
	wideAccountWidth = 30
	wideProjectWidth = 25
	wideRegionWidth  = 15
)

// PrintConfigurations prints all configurations in a formatted way.
func PrintConfigurations(configs []gcloud.Configuration, useColor bool) {
	if !useColor {
		color.NoColor = true
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow, color.Bold).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()

	for _, config := range configs {
		marker := " "
		nameColor := cyan

		if config.IsActive {
			marker = "*"
			nameColor = yellow
		}

		account := config.Properties.Core.Account
		project := config.Properties.Core.Project

		line := fmt.Sprintf("%s %s", marker, nameColor(config.Name))

		if account != "" {
			line += " " + gray(fmt.Sprintf("(%s)", account))
		}

		if project != "" {
			line += " " + gray(fmt.Sprintf("[%s]", project))
		}

		fmt.Println(line)
	}
}

// PrintCurrentConfiguration prints the current configuration name.
func PrintCurrentConfiguration(config *gcloud.Configuration, useColor bool) {
	if !useColor {
		color.NoColor = true
	}

	yellow := color.New(color.FgYellow, color.Bold).SprintFunc()
	fmt.Println(yellow(config.Name))
}

// PrintConfigurationDetails prints detailed information about a configuration.
func PrintConfigurationDetails(config *gcloud.Configuration, useColor bool) {
	if !useColor {
		color.NoColor = true
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow, color.Bold).SprintFunc()

	fmt.Printf("%s: %s\n", cyan("Configuration"), yellow(config.Name))

	if config.IsActive {
		fmt.Printf("%s: %s\n", cyan("Status"), yellow("active"))
	} else {
		fmt.Printf("%s: inactive\n", cyan("Status"))
	}

	if account := config.Properties.Core.Account; account != "" {
		fmt.Printf("%s: %s\n", cyan("Account"), account)
	}

	if project := config.Properties.Core.Project; project != "" {
		fmt.Printf("%s: %s\n", cyan("Project"), project)
	}

	if region := config.Properties.Compute.Region; region != "" {
		fmt.Printf("%s: %s\n", cyan("Region"), region)
	}

	if zone := config.Properties.Compute.Zone; zone != "" {
		fmt.Printf("%s: %s\n", cyan("Zone"), zone)
	}
}

// PrintError prints an error message.
func PrintError(message string, useColor bool) {
	if !useColor {
		color.NoColor = true
	}

	red := color.New(color.FgRed, color.Bold).SprintFunc()
	fmt.Printf("%s %s\n", red("Error:"), message)
}

// PrintSuccess prints a success message.
func PrintSuccess(message string, useColor bool) {
	if !useColor {
		color.NoColor = true
	}

	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	fmt.Printf("%s %s\n", green("Success:"), message)
}

// FormatConfigurationName formats a configuration name with marker if active.
func FormatConfigurationName(name string, isActive bool) string {
	marker := " "
	if isActive {
		marker = "*"
	}

	return fmt.Sprintf("%s %s", marker, name)
}

// TruncateString truncates a string to a maximum length.
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	if maxLen <= len(ellipsis) {
		return s[:maxLen]
	}

	return s[:maxLen-len(ellipsis)] + ellipsis
}

// AlignColumns aligns text in columns.
func AlignColumns(rows [][]string, padding int) []string {
	if len(rows) == 0 {
		return nil
	}

	colCount := len(rows[0])
	maxWidths := make([]int, colCount)

	for _, row := range rows {
		for i, cell := range row {
			cleanCell := removeANSICodes(cell)
			if len(cleanCell) > maxWidths[i] {
				maxWidths[i] = len(cleanCell)
			}
		}
	}

	result := make([]string, len(rows))
	for i, row := range rows {
		var parts []string

		for j, cell := range row {
			cleanCell := removeANSICodes(cell)

			spaces := maxWidths[j] - len(cleanCell) + padding
			if j < colCount-1 {
				parts = append(parts, cell+strings.Repeat(" ", spaces))
			} else {
				parts = append(parts, cell)
			}
		}

		result[i] = strings.Join(parts, "")
	}

	return result
}

// removeANSICodes strips SGR color sequences (ESC ... 'm') only, which is all
// fatih/color emits; other escape sequences are not handled.
func removeANSICodes(s string) string {
	inEscape := false

	var result strings.Builder

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true

			continue
		}

		if inEscape {
			if r == 'm' {
				inEscape = false
			}

			continue
		}

		result.WriteRune(r)
	}

	return result.String()
}

// ConfigOutput represents configuration data for JSON/YAML output.
type ConfigOutput struct {
	Name     string `json:"name"              yaml:"name"`
	IsActive bool   `json:"is_active"         yaml:"is_active"`
	Account  string `json:"account,omitempty" yaml:"account,omitempty"`
	Project  string `json:"project,omitempty" yaml:"project,omitempty"`
	Region   string `json:"region,omitempty"  yaml:"region,omitempty"`
	Zone     string `json:"zone,omitempty"    yaml:"zone,omitempty"`
}

// PrintConfigurationsWithFormat prints configurations in the specified format.
func PrintConfigurationsWithFormat(configs []gcloud.Configuration, format Format, useColor bool) error {
	switch format {
	case FormatJSON:
		return printConfigurationsJSON(configs)
	case FormatYAML:
		return printConfigurationsYAML(configs)
	case FormatWide:
		printConfigurationsWide(configs, useColor)

		return nil
	case FormatName:
		printConfigurationsName(configs)

		return nil
	case FormatDefault:
		PrintConfigurations(configs, useColor)

		return nil
	default:
		return fmt.Errorf("unsupported output format: %q", format)
	}
}

func toConfigOutputs(configs []gcloud.Configuration) []ConfigOutput {
	outputs := make([]ConfigOutput, len(configs))
	for i, config := range configs {
		outputs[i] = ConfigOutput{
			Name:     config.Name,
			IsActive: config.IsActive,
			Account:  config.Properties.Core.Account,
			Project:  config.Properties.Core.Project,
			Region:   config.Properties.Compute.Region,
			Zone:     config.Properties.Compute.Zone,
		}
	}

	return outputs
}

func printConfigurationsJSON(configs []gcloud.Configuration) error {
	data, err := json.MarshalIndent(toConfigOutputs(configs), "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode configurations as JSON: %w", err)
	}

	fmt.Println(string(data))

	return nil
}

func printConfigurationsYAML(configs []gcloud.Configuration) error {
	data, err := yaml.Marshal(toConfigOutputs(configs))
	if err != nil {
		return fmt.Errorf("failed to encode configurations as YAML: %w", err)
	}

	fmt.Print(string(data))

	return nil
}

func printConfigurationsWide(configs []gcloud.Configuration, useColor bool) {
	if !useColor {
		color.NoColor = true
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow, color.Bold).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	fmt.Printf(wideRowFormat,
		bold(" "),
		wideNameWidth, bold("NAME"),
		wideAccountWidth, bold("ACCOUNT"),
		wideProjectWidth, bold("PROJECT"),
		wideRegionWidth, bold("REGION"),
		bold("ZONE"))

	for _, config := range configs {
		marker := " "
		nameColor := cyan

		if config.IsActive {
			marker = "*"
			nameColor = yellow
		}

		account := config.Properties.Core.Account
		if account == "" {
			account = gray("-")
		}

		project := config.Properties.Core.Project
		if project == "" {
			project = gray("-")
		}

		region := config.Properties.Compute.Region
		if region == "" {
			region = gray("-")
		}

		zone := config.Properties.Compute.Zone
		if zone == "" {
			zone = gray("-")
		}

		fmt.Printf(wideRowFormat,
			marker,
			wideNameWidth, nameColor(TruncateString(config.Name, wideNameWidth)),
			wideAccountWidth, TruncateString(account, wideAccountWidth),
			wideProjectWidth, TruncateString(project, wideProjectWidth),
			wideRegionWidth, TruncateString(region, wideRegionWidth),
			zone)
	}
}

func printConfigurationsName(configs []gcloud.Configuration) {
	for _, config := range configs {
		fmt.Println(config.Name)
	}
}

// ValidateOutputFormat validates the output format string.
func ValidateOutputFormat(format string) (Format, error) {
	switch strings.ToLower(format) {
	case "", "default":
		return FormatDefault, nil
	case "json":
		return FormatJSON, nil
	case "yaml", "yml":
		return FormatYAML, nil
	case "wide":
		return FormatWide, nil
	case "name":
		return FormatName, nil
	default:
		return "", fmt.Errorf("unsupported output format: %s (supported: json, yaml, wide, name)", format)
	}
}
