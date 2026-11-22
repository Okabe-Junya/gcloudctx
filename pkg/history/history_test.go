package history

import (
	"os"
	"path/filepath"
	"testing"
)

// isolateHome points the user home directory at a temporary directory so the
// tests never touch the real history file.
func isolateHome(t *testing.T) string {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	return home
}

func TestGetHistoryFilePath(t *testing.T) {
	home := isolateHome(t)

	path, err := GetHistoryFilePath()
	if err != nil {
		t.Fatalf("GetHistoryFilePath failed: %v", err)
	}

	if want := filepath.Join(home, historyFileName); path != want {
		t.Errorf("GetHistoryFilePath() = %q, want %q", path, want)
	}
}

func TestSaveAndGetPreviousConfig(t *testing.T) {
	isolateHome(t)

	testConfig := "test-config"
	if err := SavePreviousConfig(testConfig); err != nil {
		t.Fatalf("SavePreviousConfig failed: %v", err)
	}

	retrieved, err := GetPreviousConfig()
	if err != nil {
		t.Fatalf("GetPreviousConfig failed: %v", err)
	}

	if retrieved != testConfig {
		t.Errorf("Expected %q, got %q", testConfig, retrieved)
	}
}

func TestGetPreviousConfigNotFound(t *testing.T) {
	isolateHome(t)

	if _, err := GetPreviousConfig(); err == nil {
		t.Error("Expected error when history file doesn't exist, got nil")
	}
}

func TestClearHistory(t *testing.T) {
	isolateHome(t)

	if err := SavePreviousConfig("test-config"); err != nil {
		t.Fatalf("SavePreviousConfig failed: %v", err)
	}

	if err := ClearHistory(); err != nil {
		t.Fatalf("ClearHistory failed: %v", err)
	}

	path, err := GetHistoryFilePath()
	if err != nil {
		t.Fatalf("GetHistoryFilePath failed: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("History file should be removed after ClearHistory")
	}
}

func TestSaveEmptyConfig(t *testing.T) {
	isolateHome(t)

	if err := SavePreviousConfig(""); err != nil {
		t.Fatalf("SavePreviousConfig failed: %v", err)
	}

	if _, err := GetPreviousConfig(); err == nil {
		t.Error("Expected error when retrieving empty config, got nil")
	}
}
