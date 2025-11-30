package interactive

import (
	"os"
	"testing"
)

func TestIsFzfInstalled(t *testing.T) {
	// This test will pass if fzf is installed on the system
	// We just check that it doesn't panic
	result := IsFzfInstalled()
	t.Logf("fzf installed: %v", result)
}

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "environment variable set",
			key:          "TEST_ENV_VAR",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "environment variable not set",
			key:          "TEST_ENV_VAR_NOT_SET",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "empty default value",
			key:          "TEST_ENV_VAR_EMPTY",
			defaultValue: "",
			envValue:     "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvOrDefault(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvOrDefault(%q, %q) = %q; want %q", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestBuildFzfArgs(t *testing.T) {
	tests := []struct {
		name        string
		envSettings map[string]string
		checkArgs   func([]string) bool
		description string
	}{
		{
			name:        "default settings",
			envSettings: map[string]string{},
			checkArgs: func(args []string) bool {
				// Should contain default height
				for i, arg := range args {
					if arg == "--height" && i+1 < len(args) {
						return args[i+1] == DefaultFzfHeight
					}
				}
				return false
			},
			description: "should use default height",
		},
		{
			name: "custom height",
			envSettings: map[string]string{
				EnvFzfHeight: "80%",
			},
			checkArgs: func(args []string) bool {
				for i, arg := range args {
					if arg == "--height" && i+1 < len(args) {
						return args[i+1] == "80%"
					}
				}
				return false
			},
			description: "should use custom height",
		},
		{
			name: "preview disabled",
			envSettings: map[string]string{
				EnvDisablePreview: "1",
			},
			checkArgs: func(args []string) bool {
				// Should not contain --preview
				for _, arg := range args {
					if arg == "--preview" {
						return false
					}
				}
				return true
			},
			description: "should not include preview",
		},
		{
			name: "preview enabled",
			envSettings: map[string]string{
				EnvDisablePreview: "0",
			},
			checkArgs: func(args []string) bool {
				// Should contain --preview
				for _, arg := range args {
					if arg == "--preview" {
						return true
					}
				}
				return false
			},
			description: "should include preview",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envSettings {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			args := buildFzfArgs("gcloudctx") // Pass dummy command path

			if !tt.checkArgs(args) {
				t.Errorf("buildFzfArgs() %s\nGot args: %v", tt.description, args)
			}
		})
	}
}

func TestBuildFzfArgsContainsRequiredOptions(t *testing.T) {
	args := buildFzfArgs("gcloudctx") // Pass dummy command path

	requiredArgs := []string{
		"--ansi",
		"--height",
		"--reverse",
		"--border",
		"--header",
		"--prompt",
	}

	for _, required := range requiredArgs {
		found := false
		for _, arg := range args {
			if arg == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("buildFzfArgs() missing required argument: %s", required)
		}
	}
}
