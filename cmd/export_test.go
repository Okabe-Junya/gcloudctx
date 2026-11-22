package cmd

import (
	"strings"
	"testing"
)

func TestMarshalExport(t *testing.T) {
	cfg := ExportConfig{Name: "prod", Project: "my-project"}

	tests := []struct {
		format string
		want   string
	}{
		{"yaml", "name: prod\nproject: my-project\n"},
		{"yml", "name: prod\nproject: my-project\n"},
		{"json", "{\n  \"name\": \"prod\",\n  \"project\": \"my-project\"\n}\n"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			got, err := marshalExport(cfg, tt.format)
			if err != nil {
				t.Fatalf("marshalExport(%q) error = %v", tt.format, err)
			}

			if string(got) != tt.want {
				t.Errorf("marshalExport(%q) = %q, want %q", tt.format, got, tt.want)
			}
		})
	}
}

func TestMarshalExportUnsupportedFormat(t *testing.T) {
	_, err := marshalExport(ExportConfig{Name: "prod"}, "toml")
	if err == nil || !strings.Contains(err.Error(), "unsupported format: toml") {
		t.Errorf("marshalExport(toml) error = %v, want unsupported format error", err)
	}
}
