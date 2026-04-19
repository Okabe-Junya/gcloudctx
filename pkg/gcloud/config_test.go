package gcloud

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGetActiveConfigurationFromList(t *testing.T) {
	tests := []struct {
		name        string
		configs     []Configuration
		wantName    string
		wantErr     bool
		errContains string
	}{
		{
			name: "single active configuration",
			configs: []Configuration{
				{Name: "default", IsActive: true},
				{Name: "production", IsActive: false},
			},
			wantName: "default",
			wantErr:  false,
		},
		{
			name: "active configuration in middle",
			configs: []Configuration{
				{Name: "dev", IsActive: false},
				{Name: "staging", IsActive: true},
				{Name: "prod", IsActive: false},
			},
			wantName: "staging",
			wantErr:  false,
		},
		{
			name:        "no active configuration",
			configs:     []Configuration{{Name: "default", IsActive: false}},
			wantErr:     true,
			errContains: "no active configuration",
		},
		{
			name:        "empty configuration list",
			configs:     []Configuration{},
			wantErr:     true,
			errContains: "no active configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetActiveConfigurationFromList(tt.configs)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetActiveConfigurationFromList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error message = %q, want to contain %q", err.Error(), tt.errContains)
				}
				return
			}
			if got != nil && got.Name != tt.wantName {
				t.Errorf("GetActiveConfigurationFromList() = %v, want %v", got.Name, tt.wantName)
			}
		})
	}
}

func TestFindConfigurationByName(t *testing.T) {
	configs := []Configuration{
		{Name: "default", IsActive: true, Properties: Properties{
			Core: CoreProperties{Account: "user@example.com", Project: "my-project"},
		}},
		{Name: "production", IsActive: false, Properties: Properties{
			Core:    CoreProperties{Account: "admin@example.com", Project: "prod-project"},
			Compute: ComputeProperties{Region: "us-central1", Zone: "us-central1-a"},
		}},
		{Name: "staging", IsActive: false},
	}

	tests := []struct {
		name        string
		configName  string
		wantFound   bool
		wantProject string
	}{
		{
			name:        "find existing configuration",
			configName:  "production",
			wantFound:   true,
			wantProject: "prod-project",
		},
		{
			name:       "find default configuration",
			configName: "default",
			wantFound:  true,
		},
		{
			name:       "configuration not found",
			configName: "nonexistent",
			wantFound:  false,
		},
		{
			name:       "empty name",
			configName: "",
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := FindConfigurationByName(configs, tt.configName)
			if found != tt.wantFound {
				t.Errorf("FindConfigurationByName() found = %v, want %v", found, tt.wantFound)
				return
			}
			if found && tt.wantProject != "" && got.Properties.Core.Project != tt.wantProject {
				t.Errorf("FindConfigurationByName() project = %v, want %v", got.Properties.Core.Project, tt.wantProject)
			}
		})
	}
}

func TestConfigurationExistsInList(t *testing.T) {
	configs := []Configuration{
		{Name: "default", IsActive: true},
		{Name: "production", IsActive: false},
		{Name: "staging", IsActive: false},
	}

	tests := []struct {
		name       string
		configName string
		want       bool
	}{
		{"exists - default", "default", true},
		{"exists - production", "production", true},
		{"not exists", "development", false},
		{"empty name", "", false},
		{"case sensitive", "Default", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConfigurationExistsInList(configs, tt.configName); got != tt.want {
				t.Errorf("ConfigurationExistsInList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateConfigurationName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple name", "myconfig", false},
		{"valid with hyphen", "my-config", false},
		{"valid with underscore", "my_config", false},
		{"valid with numbers", "config123", false},
		{"valid mixed", "my-config_123", false},
		{"empty name", "", true},
		{"starts with hyphen", "-config", true},
		{"starts with number", "123config", true},
		{"contains space", "my config", true},
		{"contains special char", "my@config", true},
		{"too long", "this-is-a-very-long-configuration-name-that-exceeds-the-maximum-allowed-length-for-gcloud-configurations", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfigurationName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfigurationName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfigurationNameBoundary(t *testing.T) {
	// Exactly 63 characters should be valid
	name63 := "a" + strings.Repeat("b", 62)
	if err := ValidateConfigurationName(name63); err != nil {
		t.Errorf("Expected 63-char name to be valid, got error: %v", err)
	}

	// 64 characters should be invalid
	name64 := "a" + strings.Repeat("b", 63)
	if err := ValidateConfigurationName(name64); err == nil {
		t.Error("Expected 64-char name to be invalid")
	}

	// Single letter should be valid
	if err := ValidateConfigurationName("a"); err != nil {
		t.Errorf("Expected single letter to be valid, got error: %v", err)
	}

	// Underscore start should be invalid
	if err := ValidateConfigurationName("_config"); err == nil {
		t.Error("Expected name starting with underscore to be invalid")
	}
}

func TestGetActiveConfigurationFromListReturnsPointerToOriginal(t *testing.T) {
	configs := []Configuration{
		{Name: "default", IsActive: true},
	}

	got, err := GetActiveConfigurationFromList(configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the returned pointer points to the original slice element
	if got.Name != "default" {
		t.Errorf("expected 'default', got %q", got.Name)
	}
}

func TestFindConfigurationByNameReturnsProperties(t *testing.T) {
	configs := []Configuration{
		{
			Name:     "full-config",
			IsActive: false,
			Properties: Properties{
				Core: CoreProperties{
					Account: "test@example.com",
					Project: "test-project",
				},
				Compute: ComputeProperties{
					Region: "asia-northeast1",
					Zone:   "asia-northeast1-a",
				},
			},
		},
	}

	got, found := FindConfigurationByName(configs, "full-config")
	if !found {
		t.Fatal("expected to find 'full-config'")
	}
	if got.Properties.Core.Account != "test@example.com" {
		t.Errorf("account = %q, want %q", got.Properties.Core.Account, "test@example.com")
	}
	if got.Properties.Compute.Region != "asia-northeast1" {
		t.Errorf("region = %q, want %q", got.Properties.Compute.Region, "asia-northeast1")
	}
	if got.Properties.Compute.Zone != "asia-northeast1-a" {
		t.Errorf("zone = %q, want %q", got.Properties.Compute.Zone, "asia-northeast1-a")
	}
}

func TestConfigurationExistsInListEmpty(t *testing.T) {
	if ConfigurationExistsInList(nil, "any") {
		t.Error("expected false for nil list")
	}
	if ConfigurationExistsInList([]Configuration{}, "any") {
		t.Error("expected false for empty list")
	}
}

func TestConfigurationJSONUnmarshal(t *testing.T) {
	// Test that Configuration struct can correctly unmarshal gcloud's JSON output
	jsonData := `[
		{
			"name": "default",
			"is_active": true,
			"properties": {
				"core": {
					"account": "user@example.com",
					"project": "my-project"
				},
				"compute": {
					"region": "us-central1",
					"zone": "us-central1-a"
				}
			}
		},
		{
			"name": "staging",
			"is_active": false,
			"properties": {
				"core": {
					"project": "staging-project"
				}
			}
		}
	]`

	var configs []Configuration
	if err := json.Unmarshal([]byte(jsonData), &configs); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if len(configs) != 2 {
		t.Fatalf("expected 2 configurations, got %d", len(configs))
	}

	// Verify first config
	if configs[0].Name != "default" {
		t.Errorf("first config name = %q, want %q", configs[0].Name, "default")
	}
	if !configs[0].IsActive {
		t.Error("expected first config to be active")
	}
	if configs[0].Properties.Core.Account != "user@example.com" {
		t.Errorf("first config account = %q, want %q", configs[0].Properties.Core.Account, "user@example.com")
	}
	if configs[0].Properties.Compute.Region != "us-central1" {
		t.Errorf("first config region = %q, want %q", configs[0].Properties.Compute.Region, "us-central1")
	}

	// Verify second config (partial properties)
	if configs[1].Name != "staging" {
		t.Errorf("second config name = %q, want %q", configs[1].Name, "staging")
	}
	if configs[1].IsActive {
		t.Error("expected second config to be inactive")
	}
	if configs[1].Properties.Core.Account != "" {
		t.Errorf("second config account should be empty, got %q", configs[1].Properties.Core.Account)
	}
	if configs[1].Properties.Core.Project != "staging-project" {
		t.Errorf("second config project = %q, want %q", configs[1].Properties.Core.Project, "staging-project")
	}
}

func TestConfigurationJSONUnmarshalStringBool(t *testing.T) {
	// gcloud may return boolean values as strings ("True"/"False") instead of native JSON booleans
	jsonData := `[
		{
			"name": "default",
			"is_active": true,
			"properties": {
				"core": {
					"account": "user@example.com",
					"project": "my-project",
					"disable_usage_reporting": "False"
				}
			}
		},
		{
			"name": "staging",
			"is_active": false,
			"properties": {
				"core": {
					"project": "staging-project",
					"disable_usage_reporting": "True"
				}
			}
		},
		{
			"name": "prod",
			"is_active": false,
			"properties": {
				"core": {
					"project": "prod-project",
					"disable_usage_reporting": true
				}
			}
		}
	]`

	var configs []Configuration
	if err := json.Unmarshal([]byte(jsonData), &configs); err != nil {
		t.Fatalf("failed to unmarshal JSON with string booleans: %v", err)
	}

	if len(configs) != 3 {
		t.Fatalf("expected 3 configurations, got %d", len(configs))
	}

	if bool(configs[0].Properties.Core.DisableUsageReport) != false {
		t.Error("expected disable_usage_reporting to be false for 'False' string")
	}
	if bool(configs[1].Properties.Core.DisableUsageReport) != true {
		t.Error("expected disable_usage_reporting to be true for 'True' string")
	}
	if bool(configs[2].Properties.Core.DisableUsageReport) != true {
		t.Error("expected disable_usage_reporting to be true for native true boolean")
	}
}

func TestMultipleActiveConfigurations(t *testing.T) {
	// gcloud should only have one active config, but test that our code returns the first one
	configs := []Configuration{
		{Name: "first", IsActive: true},
		{Name: "second", IsActive: true},
	}

	got, err := GetActiveConfigurationFromList(configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "first" {
		t.Errorf("expected first active config 'first', got %q", got.Name)
	}
}
