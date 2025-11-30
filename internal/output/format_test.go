package output

import (
	"testing"

	"github.com/Okabe-Junya/gcloudctx/pkg/gcloud"
)

func TestFormatConfigurationName(t *testing.T) {
	tests := []struct {
		name     string
		isActive bool
		expected string
	}{
		{"test-config", true, "* test-config"},
		{"test-config", false, "  test-config"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatConfigurationName(tt.name, tt.isActive)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a very long string", 10, "this is..."},
		{"exact", 5, "exact"},
		{"toolong", 5, "to..."},
		{"abc", 2, "ab"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestRemoveANSICodes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no codes", "plain text", "plain text"},
		{"with color", "\x1b[31mred text\x1b[0m", "red text"},
		{"multiple codes", "\x1b[1m\x1b[31mbold red\x1b[0m", "bold red"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeANSICodes(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestAlignColumns(t *testing.T) {
	rows := [][]string{
		{"short", "text"},
		{"very long text", "x"},
	}

	result := AlignColumns(rows, 2)
	if len(result) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(result))
	}

	// The first column should be aligned
	if len(result[0]) < len(result[1]) {
		t.Error("First row should be at least as long as second row after alignment")
	}
}

func TestPrintConfigurationsDoesNotPanic(t *testing.T) {
	configs := []gcloud.Configuration{
		{
			Name:     "test-config",
			IsActive: true,
			Properties: gcloud.Properties{
				Core: gcloud.CoreProperties{
					Account: "test@example.com",
					Project: "test-project",
				},
			},
		},
	}

	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintConfigurations panicked: %v", r)
		}
	}()

	PrintConfigurations(configs, false)
}

func TestPrintCurrentConfigurationDoesNotPanic(t *testing.T) {
	config := &gcloud.Configuration{
		Name:     "test-config",
		IsActive: true,
	}

	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintCurrentConfiguration panicked: %v", r)
		}
	}()

	PrintCurrentConfiguration(config, false)
}
