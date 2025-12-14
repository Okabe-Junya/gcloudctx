package local

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndFindLocalConfig(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "gcloudctx-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write a config file
	configName := "my-test-config"
	if err := WriteLocalConfig(tmpDir, configName); err != nil {
		t.Fatalf("WriteLocalConfig failed: %v", err)
	}

	// Verify file contents
	configPath := filepath.Join(tmpDir, ConfigFileName)
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	expected := configName + "\n"
	if string(data) != expected {
		t.Errorf("config file contents = %q, want %q", string(data), expected)
	}
}

func TestFindLocalConfigInPath(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "gcloudctx-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create nested directories
	nestedDir := filepath.Join(tmpDir, "project", "src", "deep")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	// Write config in project dir
	projectDir := filepath.Join(tmpDir, "project")
	configName := "project-config"
	if err := WriteLocalConfig(projectDir, configName); err != nil {
		t.Fatalf("WriteLocalConfig failed: %v", err)
	}

	// Find config from deep nested dir
	foundName, foundDir, err := findLocalConfigInPath(nestedDir)
	if err != nil {
		t.Fatalf("findLocalConfigInPath failed: %v", err)
	}

	if foundName != configName {
		t.Errorf("found config name = %q, want %q", foundName, configName)
	}

	if foundDir != projectDir {
		t.Errorf("found dir = %q, want %q", foundDir, projectDir)
	}
}

func TestFindLocalConfigNotFound(t *testing.T) {
	// Create a temporary directory without .gcloudctx
	tmpDir, err := os.MkdirTemp("", "gcloudctx-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	_, _, err = findLocalConfigInPath(tmpDir)
	if err == nil {
		t.Error("expected error when no config file exists")
	}
}

func TestRemoveLocalConfig(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "gcloudctx-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write a config file
	if err := WriteLocalConfig(tmpDir, "test-config"); err != nil {
		t.Fatalf("WriteLocalConfig failed: %v", err)
	}

	// Remove it
	if err := RemoveLocalConfig(tmpDir); err != nil {
		t.Fatalf("RemoveLocalConfig failed: %v", err)
	}

	// Verify it's gone
	configPath := filepath.Join(tmpDir, ConfigFileName)
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("config file should not exist after removal")
	}
}

func TestRemoveLocalConfigNonExistent(t *testing.T) {
	// Create a temporary directory without .gcloudctx
	tmpDir, err := os.MkdirTemp("", "gcloudctx-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Should not error when file doesn't exist
	if err := RemoveLocalConfig(tmpDir); err != nil {
		t.Errorf("RemoveLocalConfig should not error for non-existent file: %v", err)
	}
}

func TestFindLocalConfigEmptyFile(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "gcloudctx-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write an empty config file
	configPath := filepath.Join(tmpDir, ConfigFileName)
	if err := os.WriteFile(configPath, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to write empty config: %v", err)
	}

	_, _, err = findLocalConfigInPath(tmpDir)
	if err == nil {
		t.Error("expected error for empty config file")
	}
}

