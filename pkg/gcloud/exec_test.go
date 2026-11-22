package gcloud

import (
	"testing"
)

func TestCheckGcloudInstalled(t *testing.T) {
	err := CheckGcloudInstalled()
	if err != nil {
		t.Logf("gcloud is not installed (this is expected in some environments): %v", err)
	}
}

func TestRunGcloudCommand(t *testing.T) {
	if err := CheckGcloudInstalled(); err != nil {
		t.Skip("gcloud is not installed, skipping test")
	}

	output, err := RunGcloudCommand("version", "--format=json")
	if err != nil {
		t.Fatalf("RunGcloudCommand failed: %v", err)
	}

	if output == "" {
		t.Error("Expected non-empty output from gcloud version")
	}
}

func TestRunGcloudCommandInvalid(t *testing.T) {
	if err := CheckGcloudInstalled(); err != nil {
		t.Skip("gcloud is not installed, skipping test")
	}

	_, err := RunGcloudCommand("invalid-command-that-does-not-exist")
	if err == nil {
		t.Error("Expected error for invalid command, got nil")
	}
}
