package interactive

import (
	"testing"
)

func TestParseConfigurationName(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		{
			name:        "active configuration with account and project",
			input:       "* default (junya.okabe@hireroo.io) [core-429616]",
			expected:    "default",
			shouldError: false,
		},
		{
			name:        "non-active configuration with account and project",
			input:       "  development (dev@example.com) [dev-project-12345]",
			expected:    "development",
			shouldError: false,
		},
		{
			name:        "active configuration with only account",
			input:       "* staging (staging@example.com)",
			expected:    "staging",
			shouldError: false,
		},
		{
			name:        "non-active configuration with only project",
			input:       "  production [prod-project]",
			expected:    "production",
			shouldError: false,
		},
		{
			name:        "active configuration without account or project",
			input:       "* minimal",
			expected:    "minimal",
			shouldError: false,
		},
		{
			name:        "non-active configuration without account or project",
			input:       "  simple-config",
			expected:    "simple-config",
			shouldError: false,
		},
		{
			name:        "configuration with hyphen",
			input:       "* my-test-config (test@example.com) [test-project]",
			expected:    "my-test-config",
			shouldError: false,
		},
		{
			name:        "configuration with underscore",
			input:       "  my_test_config (test@example.com) [test-project]",
			expected:    "my_test_config",
			shouldError: false,
		},
		{
			name:        "empty line",
			input:       "",
			expected:    "",
			shouldError: true,
		},
		{
			name:        "only marker",
			input:       "*",
			expected:    "",
			shouldError: true,
		},
		{
			name:        "only parenthesized fields",
			input:       "* (account) [project]",
			expected:    "",
			shouldError: true,
		},
		{
			name:        "leading and trailing whitespace",
			input:       "   * config-name (account@example.com) [project-id]   ",
			expected:    "config-name",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseConfigurationName(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("ParseConfigurationName(%q) expected error, but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseConfigurationName(%q) unexpected error: %v", tt.input, err)
				return
			}

			if result != tt.expected {
				t.Errorf("ParseConfigurationName(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}
