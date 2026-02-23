package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/Okabe-Junya/gcloudctx/pkg/gcloud"
	"gopkg.in/yaml.v3"
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

func TestValidateOutputFormat(t *testing.T) {
	tests := []struct {
		input      string
		wantFormat Format
		wantErr    bool
	}{
		{"", FormatDefault, false},
		{"default", FormatDefault, false},
		{"json", FormatJSON, false},
		{"JSON", FormatJSON, false},
		{"yaml", FormatYAML, false},
		{"yml", FormatYAML, false},
		{"wide", FormatWide, false},
		{"name", FormatName, false},
		{"xml", "", true},
		{"csv", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ValidateOutputFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOutputFormat(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.wantFormat {
				t.Errorf("ValidateOutputFormat(%q) = %v, want %v", tt.input, got, tt.wantFormat)
			}
		})
	}
}

// captureStdout captures stdout output during the execution of f
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPrintConfigurationsWithFormatJSON(t *testing.T) {
	configs := []gcloud.Configuration{
		{
			Name:     "default",
			IsActive: true,
			Properties: gcloud.Properties{
				Core: gcloud.CoreProperties{Account: "user@example.com", Project: "my-project"},
			},
		},
		{
			Name:     "staging",
			IsActive: false,
			Properties: gcloud.Properties{
				Core:    gcloud.CoreProperties{Account: "admin@example.com", Project: "staging-project"},
				Compute: gcloud.ComputeProperties{Region: "us-central1", Zone: "us-central1-a"},
			},
		},
	}

	out := captureStdout(t, func() {
		if err := PrintConfigurationsWithFormat(configs, FormatJSON, false); err != nil {
			t.Fatalf("PrintConfigurationsWithFormat(JSON) failed: %v", err)
		}
	})

	var result []ConfigOutput
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("JSON output is not valid: %v\nOutput: %s", err, out)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 configurations, got %d", len(result))
	}
	if result[0].Name != "default" {
		t.Errorf("Expected first config name 'default', got %q", result[0].Name)
	}
	if !result[0].IsActive {
		t.Error("Expected first config to be active")
	}
	if result[1].Region != "us-central1" {
		t.Errorf("Expected second config region 'us-central1', got %q", result[1].Region)
	}
}

func TestPrintConfigurationsWithFormatYAML(t *testing.T) {
	configs := []gcloud.Configuration{
		{
			Name:     "production",
			IsActive: true,
			Properties: gcloud.Properties{
				Core: gcloud.CoreProperties{Account: "admin@example.com", Project: "prod-project"},
			},
		},
	}

	out := captureStdout(t, func() {
		if err := PrintConfigurationsWithFormat(configs, FormatYAML, false); err != nil {
			t.Fatalf("PrintConfigurationsWithFormat(YAML) failed: %v", err)
		}
	})

	var result []ConfigOutput
	if err := yaml.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("YAML output is not valid: %v\nOutput: %s", err, out)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 configuration, got %d", len(result))
	}
	if result[0].Name != "production" {
		t.Errorf("Expected config name 'production', got %q", result[0].Name)
	}
}

func TestPrintConfigurationsWithFormatName(t *testing.T) {
	configs := []gcloud.Configuration{
		{Name: "alpha", IsActive: false},
		{Name: "beta", IsActive: true},
		{Name: "gamma", IsActive: false},
	}

	out := captureStdout(t, func() {
		if err := PrintConfigurationsWithFormat(configs, FormatName, false); err != nil {
			t.Fatalf("PrintConfigurationsWithFormat(Name) failed: %v", err)
		}
	})

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d: %v", len(lines), lines)
	}
	expected := []string{"alpha", "beta", "gamma"}
	for i, want := range expected {
		if lines[i] != want {
			t.Errorf("Line %d: expected %q, got %q", i, want, lines[i])
		}
	}
}

func TestPrintConfigurationsWithFormatWide(t *testing.T) {
	configs := []gcloud.Configuration{
		{
			Name:     "dev",
			IsActive: true,
			Properties: gcloud.Properties{
				Core:    gcloud.CoreProperties{Account: "dev@example.com", Project: "dev-project"},
				Compute: gcloud.ComputeProperties{Region: "asia-northeast1", Zone: "asia-northeast1-a"},
			},
		},
	}

	out := captureStdout(t, func() {
		if err := PrintConfigurationsWithFormat(configs, FormatWide, false); err != nil {
			t.Fatalf("PrintConfigurationsWithFormat(Wide) failed: %v", err)
		}
	})

	// Wide format should contain header and data
	if !strings.Contains(out, "NAME") {
		t.Error("Wide format should contain 'NAME' header")
	}
	if !strings.Contains(out, "ACCOUNT") {
		t.Error("Wide format should contain 'ACCOUNT' header")
	}
	if !strings.Contains(out, "dev") {
		t.Error("Wide format should contain config name 'dev'")
	}
}

func TestPrintConfigurationDetailsDoesNotPanic(t *testing.T) {
	config := &gcloud.Configuration{
		Name:     "test-config",
		IsActive: true,
		Properties: gcloud.Properties{
			Core:    gcloud.CoreProperties{Account: "user@example.com", Project: "my-project"},
			Compute: gcloud.ComputeProperties{Region: "us-central1", Zone: "us-central1-a"},
		},
	}

	out := captureStdout(t, func() {
		PrintConfigurationDetails(config, false)
	})

	if !strings.Contains(out, "test-config") {
		t.Error("Details should contain the config name")
	}
	if !strings.Contains(out, "active") {
		t.Error("Details should contain status")
	}
	if !strings.Contains(out, "user@example.com") {
		t.Error("Details should contain account")
	}
}

func TestPrintErrorDoesNotPanic(t *testing.T) {
	out := captureStdout(t, func() {
		PrintError("something went wrong", false)
	})

	if !strings.Contains(out, "Error:") {
		t.Error("Error output should contain 'Error:' prefix")
	}
	if !strings.Contains(out, "something went wrong") {
		t.Error("Error output should contain the error message")
	}
}

func TestPrintSuccessDoesNotPanic(t *testing.T) {
	out := captureStdout(t, func() {
		PrintSuccess("operation completed", false)
	})

	if !strings.Contains(out, "Success:") {
		t.Error("Success output should contain 'Success:' prefix")
	}
	if !strings.Contains(out, "operation completed") {
		t.Error("Success output should contain the message")
	}
}

func TestAlignColumnsEmpty(t *testing.T) {
	result := AlignColumns(nil, 2)
	if result != nil {
		t.Errorf("Expected nil for empty input, got %v", result)
	}

	result = AlignColumns([][]string{}, 2)
	if result != nil {
		t.Errorf("Expected nil for empty slice, got %v", result)
	}
}

func TestPrintConfigurationsEmpty(t *testing.T) {
	out := captureStdout(t, func() {
		PrintConfigurations(nil, false)
	})

	if out != "" {
		t.Errorf("Expected empty output for nil configs, got %q", out)
	}
}

func TestPrintConfigurationsWithFormatDefault(t *testing.T) {
	configs := []gcloud.Configuration{
		{Name: "test", IsActive: true},
	}

	out := captureStdout(t, func() {
		if err := PrintConfigurationsWithFormat(configs, FormatDefault, false); err != nil {
			t.Fatalf("PrintConfigurationsWithFormat(Default) failed: %v", err)
		}
	})

	if !strings.Contains(out, "test") {
		t.Error("Default format should contain config name")
	}
}
