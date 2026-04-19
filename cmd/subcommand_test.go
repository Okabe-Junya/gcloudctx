package cmd

import (
	"testing"
)

func TestCreateCommandRequiresExactlyOneArg(t *testing.T) {
	cmd := createCmd
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with 0 args")
	}
	if err := cmd.Args(cmd, []string{"name"}); err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
	if err := cmd.Args(cmd, []string{"a", "b"}); err == nil {
		t.Error("expected error with 2 args")
	}
}

func TestCreateCommandFlags(t *testing.T) {
	flag := createCmd.Flags().Lookup("activate")
	if flag == nil {
		t.Fatal("expected --activate flag on create command")
	}
	if flag.DefValue != "false" {
		t.Errorf("--activate default = %q, want %q", flag.DefValue, "false")
	}
}

func TestDeleteCommandRequiresExactlyOneArg(t *testing.T) {
	cmd := deleteCmd
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with 0 args")
	}
	if err := cmd.Args(cmd, []string{"name"}); err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

func TestDeleteCommandFlags(t *testing.T) {
	flag := deleteCmd.Flags().Lookup("force")
	if flag == nil {
		t.Fatal("expected --force flag on delete command")
	}
	if flag.Shorthand != "f" {
		t.Errorf("--force shorthand = %q, want %q", flag.Shorthand, "f")
	}
}

func TestRenameCommandRequiresExactlyTwoArgs(t *testing.T) {
	cmd := renameCmd
	if err := cmd.Args(cmd, []string{"old"}); err == nil {
		t.Error("expected error with 1 arg")
	}
	if err := cmd.Args(cmd, []string{"old", "new"}); err != nil {
		t.Errorf("expected no error with 2 args, got: %v", err)
	}
}

func TestCloneCommandRequiresExactlyTwoArgs(t *testing.T) {
	cmd := cloneCmd
	if err := cmd.Args(cmd, []string{"src"}); err == nil {
		t.Error("expected error with 1 arg")
	}
	if err := cmd.Args(cmd, []string{"src", "dst"}); err != nil {
		t.Errorf("expected no error with 2 args, got: %v", err)
	}
}

func TestCloneCommandFlags(t *testing.T) {
	flag := cloneCmd.Flags().Lookup("activate")
	if flag == nil {
		t.Fatal("expected --activate flag on clone command")
	}
}

func TestUseCommandFlags(t *testing.T) {
	flags := []string{"local", "unset", "switch"}
	for _, name := range flags {
		flag := useCmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("expected --%s flag on use command", name)
		}
	}
}

func TestExportCommandFlags(t *testing.T) {
	formatFlag := exportCmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("expected --format flag on export command")
	}
	if formatFlag.Shorthand != "f" {
		t.Errorf("--format shorthand = %q, want %q", formatFlag.Shorthand, "f")
	}
	if formatFlag.DefValue != "yaml" {
		t.Errorf("--format default = %q, want %q", formatFlag.DefValue, "yaml")
	}

	outputFlag := exportCmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("expected --output flag on export command")
	}
}

func TestImportCommandFlags(t *testing.T) {
	flags := []struct {
		name     string
		defValue string
	}{
		{"activate", "false"},
		{"overwrite", "false"},
		{"name", ""},
	}

	for _, f := range flags {
		flag := importCmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("expected --%s flag on import command", f.name)
			continue
		}
		if flag.DefValue != f.defValue {
			t.Errorf("--%s default = %q, want %q", f.name, flag.DefValue, f.defValue)
		}
	}
}

func TestImportCommandRequiresExactlyOneArg(t *testing.T) {
	cmd := importCmd
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with 0 args")
	}
	if err := cmd.Args(cmd, []string{"file.yaml"}); err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

func TestCompletionCommandValidArgs(t *testing.T) {
	expected := []string{"bash", "zsh", "fish", "powershell"}
	got := completionCmd.ValidArgs

	if len(got) != len(expected) {
		t.Fatalf("ValidArgs length = %d, want %d", len(got), len(expected))
	}

	for i, v := range expected {
		if got[i] != v {
			t.Errorf("ValidArgs[%d] = %q, want %q", i, got[i], v)
		}
	}
}

func TestAutoCommandAcceptsNoArgs(t *testing.T) {
	cmd := autoCmd
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("expected no error with 0 args, got: %v", err)
	}
	if err := cmd.Args(cmd, []string{"extra"}); err == nil {
		t.Error("expected error with 1 arg")
	}
}
