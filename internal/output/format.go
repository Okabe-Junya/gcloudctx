// Package output provides formatting and display utilities for gcloudctx output.
// It handles colorized printing, configuration display, and message formatting.
package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Okabe-Junya/gcloudctx/pkg/gcloud"
	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

// OutputFormat represents the output format type
type OutputFormat string

const (
	FormatDefault OutputFormat = ""
	FormatJSON    OutputFormat = "json"
	FormatYAML    OutputFormat = "yaml"
	FormatWide    OutputFormat = "wide"
	FormatName    OutputFormat = "name"
)

// PrintConfigurations prints all configurations in a formatted way
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

		// Format: * name (account) [project]
		line := fmt.Sprintf("%s %s", marker, nameColor(config.Name))

		if account != "" {
			line += fmt.Sprintf(" %s", gray(fmt.Sprintf("(%s)", account)))
		}
		if project != "" {
			line += fmt.Sprintf(" %s", gray(fmt.Sprintf("[%s]", project)))
		}

		fmt.Println(line)
	}
}

// PrintCurrentConfiguration prints the current configuration name
func PrintCurrentConfiguration(config *gcloud.Configuration, useColor bool) {
	if !useColor {
		color.NoColor = true
	}

	yellow := color.New(color.FgYellow, color.Bold).SprintFunc()
	fmt.Println(yellow(config.Name))
}

// PrintConfigurationDetails prints detailed information about a configuration
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

// PrintError prints an error message
func PrintError(message string, useColor bool) {
	if !useColor {
		color.NoColor = true
	}

	red := color.New(color.FgRed, color.Bold).SprintFunc()
	fmt.Printf("%s %s\n", red("Error:"), message)
}

// PrintSuccess prints a success message
func PrintSuccess(message string, useColor bool) {
	if !useColor {
		color.NoColor = true
	}

	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	fmt.Printf("%s %s\n", green("Success:"), message)
}

// FormatConfigurationName formats a configuration name with marker if active
func FormatConfigurationName(name string, isActive bool) string {
	marker := " "
	if isActive {
		marker = "*"
	}
	return fmt.Sprintf("%s %s", marker, name)
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// AlignColumns aligns text in columns
func AlignColumns(rows [][]string, padding int) []string {
	if len(rows) == 0 {
		return nil
	}

	// Find max width for each column
	colCount := len(rows[0])
	maxWidths := make([]int, colCount)

	for _, row := range rows {
		for i, cell := range row {
			// Remove ANSI color codes for width calculation
			cleanCell := removeANSICodes(cell)
			if len(cleanCell) > maxWidths[i] {
				maxWidths[i] = len(cleanCell)
			}
		}
	}

	// Format rows
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

// removeANSICodes removes ANSI color codes from a string
func removeANSICodes(s string) string {
	// Simple ANSI code removal (may not handle all cases)
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

// ConfigOutput represents configuration data for JSON/YAML output
type ConfigOutput struct {
	Name     string `json:"name" yaml:"name"`
	IsActive bool   `json:"is_active" yaml:"is_active"`
	Account  string `json:"account,omitempty" yaml:"account,omitempty"`
	Project  string `json:"project,omitempty" yaml:"project,omitempty"`
	Region   string `json:"region,omitempty" yaml:"region,omitempty"`
	Zone     string `json:"zone,omitempty" yaml:"zone,omitempty"`
}

// PrintConfigurationsWithFormat prints configurations in the specified format
func PrintConfigurationsWithFormat(configs []gcloud.Configuration, format OutputFormat, useColor bool) error {
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
	default:
		PrintConfigurations(configs, useColor)
		return nil
	}
}

func printConfigurationsJSON(configs []gcloud.Configuration) error {
	output := make([]ConfigOutput, len(configs))
	for i, c := range configs {
		output[i] = ConfigOutput{
			Name:     c.Name,
			IsActive: c.IsActive,
			Account:  c.Properties.Core.Account,
			Project:  c.Properties.Core.Project,
			Region:   c.Properties.Compute.Region,
			Zone:     c.Properties.Compute.Zone,
		}
	}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func printConfigurationsYAML(configs []gcloud.Configuration) error {
	output := make([]ConfigOutput, len(configs))
	for i, c := range configs {
		output[i] = ConfigOutput{
			Name:     c.Name,
			IsActive: c.IsActive,
			Account:  c.Properties.Core.Account,
			Project:  c.Properties.Core.Project,
			Region:   c.Properties.Compute.Region,
			Zone:     c.Properties.Compute.Zone,
		}
	}
	data, err := yaml.Marshal(output)
	if err != nil {
		return err
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

	// Print header
	fmt.Printf("%s  %-20s  %-30s  %-25s  %-15s  %s\n",
		bold(" "),
		bold("NAME"),
		bold("ACCOUNT"),
		bold("PROJECT"),
		bold("REGION"),
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

		fmt.Printf("%s  %-20s  %-30s  %-25s  %-15s  %s\n",
			marker,
			nameColor(TruncateString(config.Name, 20)),
			TruncateString(account, 30),
			TruncateString(project, 25),
			TruncateString(region, 15),
			zone)
	}
}

func printConfigurationsName(configs []gcloud.Configuration) {
	for _, config := range configs {
		fmt.Println(config.Name)
	}
}

// ValidateOutputFormat validates the output format string
func ValidateOutputFormat(format string) (OutputFormat, error) {
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
