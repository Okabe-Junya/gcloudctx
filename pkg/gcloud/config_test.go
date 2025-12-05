package gcloud

import (
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
			got, err := getActiveConfigurationFromList(tt.configs)
			if (err != nil) != tt.wantErr {
				t.Errorf("getActiveConfigurationFromList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" {
				if !containsString(err.Error(), tt.errContains) {
					t.Errorf("error message = %q, want to contain %q", err.Error(), tt.errContains)
				}
				return
			}
			if got != nil && got.Name != tt.wantName {
				t.Errorf("getActiveConfigurationFromList() = %v, want %v", got.Name, tt.wantName)
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
			got, found := findConfigurationByName(configs, tt.configName)
			if found != tt.wantFound {
				t.Errorf("findConfigurationByName() found = %v, want %v", found, tt.wantFound)
				return
			}
			if found && tt.wantProject != "" && got.Properties.Core.Project != tt.wantProject {
				t.Errorf("findConfigurationByName() project = %v, want %v", got.Properties.Core.Project, tt.wantProject)
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
			if got := configurationExistsInList(configs, tt.configName); got != tt.want {
				t.Errorf("configurationExistsInList() = %v, want %v", got, tt.want)
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

// containsString checks if s contains substr
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
