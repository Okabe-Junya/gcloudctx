package history

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetHistoryFilePath(t *testing.T) {
	path, err := GetHistoryFilePath()
	if err != nil {
		t.Fatalf("GetHistoryFilePath failed: %v", err)
	}
	if path == "" {
		t.Error("Expected non-empty history file path")
	}
	if !filepath.IsAbs(path) {
		t.Error("Expected absolute path")
	}
}

func TestSaveAndGetPreviousConfig(t *testing.T) {
	// This test uses the actual history file, so we need to clean up
	path, err := GetHistoryFilePath()
	if err != nil {
		t.Fatalf("GetHistoryFilePath failed: %v", err)
	}

	// Save original state
	originalContent, _ := os.ReadFile(path)
	defer func() {
		if originalContent != nil {
			os.WriteFile(path, originalContent, 0600)
		} else {
			os.Remove(path)
		}
	}()

	// Test saving a configuration
	testConfig := "test-config"
	err = SavePreviousConfig(testConfig)
	if err != nil {
		t.Fatalf("SavePreviousConfig failed: %v", err)
	}

	// Test retrieving the configuration
	retrieved, err := GetPreviousConfig()
	if err != nil {
		t.Fatalf("GetPreviousConfig failed: %v", err)
	}

	if retrieved != testConfig {
		t.Errorf("Expected %q, got %q", testConfig, retrieved)
	}
}

func TestGetPreviousConfigNotFound(t *testing.T) {
	path, err := GetHistoryFilePath()
	if err != nil {
		t.Fatalf("GetHistoryFilePath failed: %v", err)
	}

	// Save original state and remove file
	originalContent, _ := os.ReadFile(path)
	os.Remove(path)
	defer func() {
		if originalContent != nil {
			os.WriteFile(path, originalContent, 0600)
		}
	}()

	// Test retrieving when file doesn't exist
	_, err = GetPreviousConfig()
	if err == nil {
		t.Error("Expected error when history file doesn't exist, got nil")
	}
}

func TestClearHistory(t *testing.T) {
	path, err := GetHistoryFilePath()
	if err != nil {
		t.Fatalf("GetHistoryFilePath failed: %v", err)
	}

	// Save original state
	originalContent, _ := os.ReadFile(path)
	defer func() {
		if originalContent != nil {
			os.WriteFile(path, originalContent, 0600)
		}
	}()

	// Save a configuration
	err = SavePreviousConfig("test-config")
	if err != nil {
		t.Fatalf("SavePreviousConfig failed: %v", err)
	}

	// Clear history
	err = ClearHistory()
	if err != nil {
		t.Fatalf("ClearHistory failed: %v", err)
	}

	// Verify file is removed
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("History file should be removed after ClearHistory")
	}
}

func TestSaveEmptyConfig(t *testing.T) {
	path, err := GetHistoryFilePath()
	if err != nil {
		t.Fatalf("GetHistoryFilePath failed: %v", err)
	}

	// Save original state
	originalContent, _ := os.ReadFile(path)
	defer func() {
		if originalContent != nil {
			os.WriteFile(path, originalContent, 0600)
		} else {
			os.Remove(path)
		}
	}()

	// Save empty string
	err = SavePreviousConfig("")
	if err != nil {
		t.Fatalf("SavePreviousConfig failed: %v", err)
	}

	// Try to retrieve - should return error for empty config
	_, err = GetPreviousConfig()
	if err == nil {
		t.Error("Expected error when retrieving empty config, got nil")
	}
}
