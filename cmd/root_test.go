package cmd

import (
	"strings"
	"testing"
)

func TestBuildVersionString(t *testing.T) {
	tests := []struct {
		name    string
		version string
		commit  string
		date    string
		want    string
	}{
		{
			name:    "dev defaults",
			version: "dev",
			commit:  "none",
			date:    "unknown",
			want:    "dev",
		},
		{
			name:    "release with full commit",
			version: "v1.2.3",
			commit:  "abc1234def5678",
			date:    "2025-01-01",
			want:    "v1.2.3 (commit: abc1234) built at 2025-01-01",
		},
		{
			name:    "short commit hash",
			version: "v0.1.0",
			commit:  "abc",
			date:    "2025-06-01",
			want:    "v0.1.0 (commit: abc) built at 2025-06-01",
		},
		{
			name:    "no date",
			version: "v1.0.0",
			commit:  "1234567",
			date:    "unknown",
			want:    "v1.0.0 (commit: 1234567)",
		},
		{
			name:    "no commit",
			version: "v1.0.0",
			commit:  "none",
			date:    "2025-01-01",
			want:    "v1.0.0 built at 2025-01-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore package-level vars
			origVersion, origCommit, origDate := Version, Commit, Date
			defer func() { Version, Commit, Date = origVersion, origCommit, origDate }()

			Version = tt.version
			Commit = tt.commit
			Date = tt.date

			got := buildVersionString()
			if got != tt.want {
				t.Errorf("buildVersionString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRootCommandStructure(t *testing.T) {
	// Verify the root command has expected subcommands
	expectedSubcommands := []string{
		"create", "delete", "rename", "clone",
		"use", "auto", "export", "import",
		"completion",
	}

	subcommands := rootCmd.Commands()
	subcommandNames := make(map[string]bool)
	for _, cmd := range subcommands {
		subcommandNames[cmd.Name()] = true
	}

	for _, expected := range expectedSubcommands {
		if !subcommandNames[expected] {
			t.Errorf("expected subcommand %q not found on root command", expected)
		}
	}
}

func TestRootCommandFlags(t *testing.T) {
	flags := []struct {
		name      string
		shorthand string
	}{
		{"list", "l"},
		{"current", "c"},
		{"interactive", "i"},
		{"sync-adc", ""},
		{"impersonate-service-account", ""},
		{"info", ""},
		{"no-color", ""},
		{"output", "o"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := rootCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("expected flag %q not found", f.name)
				return
			}
			if f.shorthand != "" && flag.Shorthand != f.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", f.name, flag.Shorthand, f.shorthand)
			}
		})
	}
}

func TestRootCommandAcceptsMaxOneArg(t *testing.T) {
	// rootCmd.Args is cobra.MaximumNArgs(1)
	// Validate by checking that 2 args fails
	cmd := rootCmd
	cmd.SetArgs([]string{"arg1", "arg2"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error with 2 args, got nil")
	}
	if !strings.Contains(err.Error(), "accepts at most 1 arg") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestPreviewCommandIsHidden(t *testing.T) {
	subcommands := rootCmd.Commands()
	for _, cmd := range subcommands {
		if cmd.Name() == "__preview" {
			if !cmd.Hidden {
				t.Error("preview command should be hidden")
			}
			return
		}
	}
	t.Error("preview command not found")
}
