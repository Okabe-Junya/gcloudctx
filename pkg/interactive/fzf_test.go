package interactive

import (
	"slices"
	"testing"
)

func TestIsFzfInstalled(t *testing.T) {
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
				t.Setenv(tt.key, tt.envValue)
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
				return !slices.Contains(args, "--preview")
			},
			description: "should not include preview",
		},
		{
			name: "preview enabled",
			envSettings: map[string]string{
				EnvDisablePreview: "0",
			},
			checkArgs: func(args []string) bool {
				return slices.Contains(args, "--preview")
			},
			description: "should include preview",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envSettings {
				t.Setenv(key, value)
			}

			args := buildFzfArgs("gcloudctx")

			if !tt.checkArgs(args) {
				t.Errorf("buildFzfArgs() %s\nGot args: %v", tt.description, args)
			}
		})
	}
}

func TestBuildFzfArgsContainsRequiredOptions(t *testing.T) {
	args := buildFzfArgs("gcloudctx")

	requiredArgs := []string{
		"--ansi",
		"--height",
		"--reverse",
		"--border",
		"--header",
		"--prompt",
	}

	for _, required := range requiredArgs {
		found := slices.Contains(args, required)

		if !found {
			t.Errorf("buildFzfArgs() missing required argument: %s", required)
		}
	}
}
