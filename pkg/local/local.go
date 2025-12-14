// Package local provides functionality for directory-based configuration management.
// It allows users to associate a gcloud configuration with a specific directory
// using a .gcloudctx file.
package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfigFileName is the name of the local configuration file
const ConfigFileName = ".gcloudctx"

// FindLocalConfig searches for a .gcloudctx file starting from the current directory
// and walking up to the root. Returns the configuration name and the directory where
// it was found, or an error if not found.
func FindLocalConfig() (configName string, dir string, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to get current directory: %w", err)
	}

	return findLocalConfigInPath(cwd)
}

// findLocalConfigInPath searches for .gcloudctx file starting from the given path
func findLocalConfigInPath(startPath string) (configName string, dir string, err error) {
	dir = startPath

	for {
		configPath := filepath.Join(dir, ConfigFileName)
		if _, err := os.Stat(configPath); err == nil {
			// Found the file, read its contents
			data, err := os.ReadFile(configPath)
			if err != nil {
				return "", "", fmt.Errorf("failed to read %s: %w", configPath, err)
			}

			name := strings.TrimSpace(string(data))
			if name == "" {
				return "", "", fmt.Errorf("%s is empty", configPath)
			}

			return name, dir, nil
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return "", "", fmt.Errorf("no %s file found", ConfigFileName)
}

// WriteLocalConfig writes a configuration name to a .gcloudctx file in the specified directory
func WriteLocalConfig(dir, configName string) error {
	configPath := filepath.Join(dir, ConfigFileName)
	if err := os.WriteFile(configPath, []byte(configName+"\n"), 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", configPath, err)
	}
	return nil
}

// WriteLocalConfigCurrent writes a configuration name to .gcloudctx in the current directory
func WriteLocalConfigCurrent(configName string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	return WriteLocalConfig(cwd, configName)
}

// RemoveLocalConfig removes the .gcloudctx file from the specified directory
func RemoveLocalConfig(dir string) error {
	configPath := filepath.Join(dir, ConfigFileName)
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove %s: %w", configPath, err)
	}
	return nil
}

// RemoveLocalConfigCurrent removes the .gcloudctx file from the current directory
func RemoveLocalConfigCurrent() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	return RemoveLocalConfig(cwd)
}

// GetLocalConfigPath returns the path to the .gcloudctx file in the current directory
func GetLocalConfigPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	return filepath.Join(cwd, ConfigFileName), nil
}

// LocalConfigExists checks if a .gcloudctx file exists in the current directory
func LocalConfigExists() bool {
	path, err := GetLocalConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

