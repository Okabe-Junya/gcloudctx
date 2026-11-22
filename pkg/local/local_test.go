package local

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndFindLocalConfig(t *testing.T) {
	tmpDir := t.TempDir()

	configName := "my-test-config"
	if err := WriteLocalConfig(tmpDir, configName); err != nil {
		t.Fatalf("WriteLocalConfig failed: %v", err)
	}

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
	tmpDir := t.TempDir()

	nestedDir := filepath.Join(tmpDir, "project", "src", "deep")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	projectDir := filepath.Join(tmpDir, "project")

	configName := "project-config"
	if err := WriteLocalConfig(projectDir, configName); err != nil {
		t.Fatalf("WriteLocalConfig failed: %v", err)
	}

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
	tmpDir := t.TempDir()

	_, _, err := findLocalConfigInPath(tmpDir)
	if err == nil {
		t.Error("expected error when no config file exists")
	}
}

func TestRemoveLocalConfig(t *testing.T) {
	tmpDir := t.TempDir()

	if err := WriteLocalConfig(tmpDir, "test-config"); err != nil {
		t.Fatalf("WriteLocalConfig failed: %v", err)
	}

	if err := RemoveLocalConfig(tmpDir); err != nil {
		t.Fatalf("RemoveLocalConfig failed: %v", err)
	}

	configPath := filepath.Join(tmpDir, ConfigFileName)
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("config file should not exist after removal")
	}
}

func TestRemoveLocalConfigNonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	if err := RemoveLocalConfig(tmpDir); err != nil {
		t.Errorf("RemoveLocalConfig should not error for non-existent file: %v", err)
	}
}

func TestFindLocalConfigEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	configPath := filepath.Join(tmpDir, ConfigFileName)
	if err := os.WriteFile(configPath, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to write empty config: %v", err)
	}

	_, _, err := findLocalConfigInPath(tmpDir)
	if err == nil {
		t.Error("expected error for empty config file")
	}
}

func TestFindLocalConfigWhitespaceOnly(t *testing.T) {
	tmpDir := t.TempDir()

	configPath := filepath.Join(tmpDir, ConfigFileName)
	if err := os.WriteFile(configPath, []byte("   \n\t  \n"), 0o644); err != nil {
		t.Fatalf("failed to write whitespace config: %v", err)
	}

	_, _, err := findLocalConfigInPath(tmpDir)
	if err == nil {
		t.Error("expected error for whitespace-only config file")
	}
}

func TestWriteLocalConfigOverwrite(t *testing.T) {
	tmpDir := t.TempDir()

	if err := WriteLocalConfig(tmpDir, "first-config"); err != nil {
		t.Fatalf("WriteLocalConfig failed: %v", err)
	}

	if err := WriteLocalConfig(tmpDir, "second-config"); err != nil {
		t.Fatalf("WriteLocalConfig overwrite failed: %v", err)
	}

	name, _, err := findLocalConfigInPath(tmpDir)
	if err != nil {
		t.Fatalf("findLocalConfigInPath failed: %v", err)
	}

	if name != "second-config" {
		t.Errorf("expected 'second-config', got %q", name)
	}
}

func TestWriteLocalConfigPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	if err := WriteLocalConfig(tmpDir, "test-config"); err != nil {
		t.Fatalf("WriteLocalConfig failed: %v", err)
	}

	configPath := filepath.Join(tmpDir, ConfigFileName)

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("failed to stat config file: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("expected permissions 0600, got %o", perm)
	}
}

func TestFindLocalConfigTraversesUp(t *testing.T) {
	tmpDir := t.TempDir()

	deepDir := filepath.Join(tmpDir, "a", "b", "c", "d", "e")
	if err := os.MkdirAll(deepDir, 0o755); err != nil {
		t.Fatalf("failed to create deep dir: %v", err)
	}

	if err := WriteLocalConfig(tmpDir, "root-config"); err != nil {
		t.Fatalf("WriteLocalConfig failed: %v", err)
	}

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
	tmpDir := t.TempDir()

	childDir := filepath.Join(tmpDir, "child")
	if err := os.MkdirAll(childDir, 0o755); err != nil {
		t.Fatalf("failed to create child dir: %v", err)
	}

	if err := WriteLocalConfig(tmpDir, "parent-config"); err != nil {
		t.Fatalf("WriteLocalConfig parent failed: %v", err)
	}

	if err := WriteLocalConfig(childDir, "child-config"); err != nil {
		t.Fatalf("WriteLocalConfig child failed: %v", err)
	}

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
	tmpDir := t.TempDir()

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
