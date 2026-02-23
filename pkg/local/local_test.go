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

func TestFindLocalConfigWhitespaceOnly(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gcloudctx-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write a config file with only whitespace
	configPath := filepath.Join(tmpDir, ConfigFileName)
	if err := os.WriteFile(configPath, []byte("   \n\t  \n"), 0o644); err != nil {
		t.Fatalf("failed to write whitespace config: %v", err)
	}

	_, _, err = findLocalConfigInPath(tmpDir)
	if err == nil {
		t.Error("expected error for whitespace-only config file")
	}
}

func TestWriteLocalConfigOverwrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gcloudctx-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write first config
	if err := WriteLocalConfig(tmpDir, "first-config"); err != nil {
		t.Fatalf("WriteLocalConfig failed: %v", err)
	}

	// Overwrite with second config
	if err := WriteLocalConfig(tmpDir, "second-config"); err != nil {
		t.Fatalf("WriteLocalConfig overwrite failed: %v", err)
	}

	// Verify the overwritten content
	name, _, err := findLocalConfigInPath(tmpDir)
	if err != nil {
		t.Fatalf("findLocalConfigInPath failed: %v", err)
	}
	if name != "second-config" {
		t.Errorf("expected 'second-config', got %q", name)
	}
}

func TestWriteLocalConfigPermissions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gcloudctx-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := WriteLocalConfig(tmpDir, "test-config"); err != nil {
		t.Fatalf("WriteLocalConfig failed: %v", err)
	}

	configPath := filepath.Join(tmpDir, ConfigFileName)
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("failed to stat config file: %v", err)
	}

	// Check that the file has restrictive permissions (0600)
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("expected permissions 0600, got %o", perm)
	}
}

func TestFindLocalConfigTraversesUp(t *testing.T) {
	// Create a deep directory structure with config at root
	tmpDir, err := os.MkdirTemp("", "gcloudctx-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	deepDir := filepath.Join(tmpDir, "a", "b", "c", "d", "e")
	if err := os.MkdirAll(deepDir, 0o755); err != nil {
		t.Fatalf("failed to create deep dir: %v", err)
	}

	// Place config at root
	if err := WriteLocalConfig(tmpDir, "root-config"); err != nil {
		t.Fatalf("WriteLocalConfig failed: %v", err)
	}

	// Should find it from deep directory
	name, dir, err := findLocalConfigInPath(deepDir)
	if err != nil {
		t.Fatalf("findLocalConfigInPath failed: %v", err)
	}
	if name != "root-config" {
		t.Errorf("expected 'root-config', got %q", name)
	}
	if dir != tmpDir {
		t.Errorf("expected dir %q, got %q", tmpDir, dir)
	}
}

func TestFindLocalConfigClosestWins(t *testing.T) {
	// Config at parent and child - child should win
	tmpDir, err := os.MkdirTemp("", "gcloudctx-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	childDir := filepath.Join(tmpDir, "child")
	if err := os.MkdirAll(childDir, 0o755); err != nil {
		t.Fatalf("failed to create child dir: %v", err)
	}

	// Place configs at both levels
	if err := WriteLocalConfig(tmpDir, "parent-config"); err != nil {
		t.Fatalf("WriteLocalConfig parent failed: %v", err)
	}
	if err := WriteLocalConfig(childDir, "child-config"); err != nil {
		t.Fatalf("WriteLocalConfig child failed: %v", err)
	}

	// From child dir, should find child config
	name, dir, err := findLocalConfigInPath(childDir)
	if err != nil {
		t.Fatalf("findLocalConfigInPath failed: %v", err)
	}
	if name != "child-config" {
		t.Errorf("expected 'child-config', got %q", name)
	}
	if dir != childDir {
		t.Errorf("expected dir %q, got %q", childDir, dir)
	}
}

func TestWriteLocalConfigInvalidDir(t *testing.T) {
	err := WriteLocalConfig("/nonexistent/path/that/doesnt/exist", "test")
	if err == nil {
		t.Error("expected error when writing to nonexistent directory")
	}
}

func TestFindLocalConfigWithTrailingNewline(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gcloudctx-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// WriteLocalConfig adds a newline, verify it's trimmed when read
	if err := WriteLocalConfig(tmpDir, "my-config"); err != nil {
		t.Fatalf("WriteLocalConfig failed: %v", err)
	}

	name, _, err := findLocalConfigInPath(tmpDir)
	if err != nil {
		t.Fatalf("findLocalConfigInPath failed: %v", err)
	}
	if name != "my-config" {
		t.Errorf("expected 'my-config', got %q (possible untrimmed newline)", name)
	}
}
