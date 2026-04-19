package gcloud

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ListConfigurations returns all available gcloud configurations
func ListConfigurations() ([]Configuration, error) {
	output, err := RunGcloudCommand("config", "configurations", "list", "--format=json")
	if err != nil {
		return nil, fmt.Errorf("failed to list configurations: %w", err)
	}

	var configs []Configuration
	if err := json.Unmarshal([]byte(output), &configs); err != nil {
		return nil, fmt.Errorf("failed to parse configurations: %w", err)
	}

	return configs, nil
}

// GetActiveConfiguration returns the currently active configuration
func GetActiveConfiguration() (*Configuration, error) {
	configs, err := ListConfigurations()
	if err != nil {
		return nil, err
	}

	for _, config := range configs {
		if config.IsActive {
			return &config, nil
		}
	}

	return nil, fmt.Errorf("no active configuration found")
}

// ActivateConfiguration activates a specific configuration
func ActivateConfiguration(name string) error {
	if err := RunGcloudCommandQuiet("config", "configurations", "activate", name); err != nil {
		return fmt.Errorf("failed to activate configuration %q: %w", name, err)
	}
	return nil
}

// ConfigurationExists checks if a configuration exists
func ConfigurationExists(name string) bool {
	configs, err := ListConfigurations()
	if err != nil {
		return false
	}

	for _, config := range configs {
		if config.Name == name {
			return true
		}
	}

	return false
}

// SyncADC synchronizes Application Default Credentials with the current configuration
func SyncADC(impersonateServiceAccount string) error {
	args := []string{"auth", "application-default", "login"}

	if impersonateServiceAccount != "" {
		args = append(args, "--impersonate-service-account", impersonateServiceAccount)
	}

	if err := RunGcloudCommandInteractive(args...); err != nil {
		return fmt.Errorf("failed to sync ADC: %w", err)
	}

	return nil
}

// GetConfigurationInfo returns detailed information about a configuration
func GetConfigurationInfo(name string) (*Configuration, error) {
	configs, err := ListConfigurations()
	if err != nil {
		return nil, err
	}

	for _, config := range configs {
		if config.Name == name {
			return &config, nil
		}
	}

	return nil, fmt.Errorf("configuration %q not found", name)
}

// GetCurrentProject returns the current project from active configuration
func GetCurrentProject() (string, error) {
	output, err := RunGcloudCommand("config", "get-value", "project")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// GetCurrentAccount returns the current account from active configuration
func GetCurrentAccount() (string, error) {
	output, err := RunGcloudCommand("config", "get-value", "account")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// CreateConfiguration creates a new gcloud configuration
func CreateConfiguration(name string) error {
	if ConfigurationExists(name) {
		return fmt.Errorf("configuration %q already exists", name)
	}

	if err := RunGcloudCommandQuiet("config", "configurations", "create", name); err != nil {
		return fmt.Errorf("failed to create configuration %q: %w", name, err)
	}
	return nil
}

// DeleteConfiguration deletes a gcloud configuration
func DeleteConfiguration(name string) error {
	if !ConfigurationExists(name) {
		return fmt.Errorf("configuration %q does not exist", name)
	}

	// Get active configuration to prevent deletion
	activeConfig, err := GetActiveConfiguration()
	if err != nil {
		return err
	}

	if activeConfig.Name == name {
		return fmt.Errorf("cannot delete active configuration %q", name)
	}

	if err := RunGcloudCommandQuiet("config", "configurations", "delete", name, "--quiet"); err != nil {
		return fmt.Errorf("failed to delete configuration %q: %w", name, err)
	}
	return nil
}

// CloneConfiguration creates a new configuration by copying properties from an existing one
func CloneConfiguration(sourceName, targetName string) error {
	if !ConfigurationExists(sourceName) {
		return fmt.Errorf("source configuration %q does not exist", sourceName)
	}

	if ConfigurationExists(targetName) {
		return fmt.Errorf("target configuration %q already exists", targetName)
	}

	// Get the source configuration details
	sourceConfig, err := GetConfigurationInfo(sourceName)
	if err != nil {
		return err
	}

	// Create the new configuration
	if err := CreateConfiguration(targetName); err != nil {
		return err
	}

	// Copy properties to new configuration
	if err := copyConfigProperties(sourceConfig, targetName); err != nil {
		// Clean up on failure
		if cleanupErr := cleanupConfiguration(targetName); cleanupErr != nil {
			return fmt.Errorf("failed to copy properties: %w (cleanup also failed: %v)", err, cleanupErr)
		}
		return fmt.Errorf("failed to copy properties: %w", err)
	}

	return nil
}

// RenameConfiguration renames a gcloud configuration
func RenameConfiguration(oldName, newName string) error {
	if !ConfigurationExists(oldName) {
		return fmt.Errorf("configuration %q does not exist", oldName)
	}

	if ConfigurationExists(newName) {
		return fmt.Errorf("configuration %q already exists", newName)
	}

	// Get the old configuration details before starting
	oldConfig, err := GetConfigurationInfo(oldName)
	if err != nil {
		return err
	}

	// gcloud doesn't have a rename command, so we need to create a new one
	// and copy the properties
	if err := CreateConfiguration(newName); err != nil {
		return err
	}

	// Copy properties to new configuration
	if err := copyConfigProperties(oldConfig, newName); err != nil {
		// Clean up on failure
		if cleanupErr := cleanupConfiguration(newName); cleanupErr != nil {
			return fmt.Errorf("failed to copy properties: %w (cleanup also failed: %v)", err, cleanupErr)
		}
		return fmt.Errorf("failed to copy properties: %w", err)
	}

	// If old config was active, switch to new one
	if oldConfig.IsActive {
		if err := ActivateConfiguration(newName); err != nil {
			// Clean up on failure
			if cleanupErr := cleanupConfiguration(newName); cleanupErr != nil {
				return fmt.Errorf("failed to activate configuration: %w (cleanup also failed: %v)", err, cleanupErr)
			}
			return fmt.Errorf("failed to activate configuration: %w", err)
		}
	}

	// Delete old configuration
	if err := DeleteConfiguration(oldName); err != nil {
		return fmt.Errorf("failed to delete old configuration %q: %w", oldName, err)
	}

	return nil
}

// SetConfigProperties sets the account, project, region, and zone properties on a target configuration.
// Empty values are skipped.
func SetConfigProperties(targetName, account, project, region, zone string) error {
	props := []struct {
		key   string
		value string
	}{
		{"account", account},
		{"project", project},
		{"compute/region", region},
		{"compute/zone", zone},
	}

	for _, p := range props {
		if p.value == "" {
			continue
		}
		if err := RunGcloudCommandQuiet("config", "set", p.key, p.value, "--configuration", targetName); err != nil {
			return fmt.Errorf("failed to set %s property: %w", p.key, err)
		}
	}

	return nil
}

// copyConfigProperties copies properties from one configuration to another
func copyConfigProperties(source *Configuration, targetName string) error {
	return SetConfigProperties(
		targetName,
		source.Properties.Core.Account,
		source.Properties.Core.Project,
		source.Properties.Compute.Region,
		source.Properties.Compute.Zone,
	)
}

// cleanupConfiguration attempts to delete a configuration and returns any error encountered
func cleanupConfiguration(name string) error {
	if err := DeleteConfiguration(name); err != nil {
		return fmt.Errorf("failed to cleanup configuration %q: %w", name, err)
	}
	return nil
}

// GetActiveConfigurationFromList finds the active configuration from a list
// This is a pure function for easier testing
func GetActiveConfigurationFromList(configs []Configuration) (*Configuration, error) {
	for i := range configs {
		if configs[i].IsActive {
			return &configs[i], nil
		}
	}
	return nil, fmt.Errorf("no active configuration found")
}

// FindConfigurationByName finds a configuration by name from a list
// Returns the configuration and a boolean indicating if it was found
func FindConfigurationByName(configs []Configuration, name string) (*Configuration, bool) {
	for i := range configs {
		if configs[i].Name == name {
			return &configs[i], true
		}
	}
	return nil, false
}

// ConfigurationExistsInList checks if a configuration exists in a list
func ConfigurationExistsInList(configs []Configuration, name string) bool {
	_, found := FindConfigurationByName(configs, name)
	return found
}

// configNameRegex validates configuration names
// Must start with a letter, contain only alphanumeric, hyphens, and underscores
var configNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

// MaxConfigNameLength is the maximum allowed length for a configuration name
const MaxConfigNameLength = 63

// ValidateConfigurationName validates a configuration name
func ValidateConfigurationName(name string) error {
	if name == "" {
		return fmt.Errorf("configuration name cannot be empty")
	}

	if len(name) > MaxConfigNameLength {
		return fmt.Errorf("configuration name cannot exceed %d characters", MaxConfigNameLength)
	}

	if !configNameRegex.MatchString(name) {
		return fmt.Errorf("configuration name must start with a letter and contain only alphanumeric characters, hyphens, and underscores")
	}

	return nil
}
