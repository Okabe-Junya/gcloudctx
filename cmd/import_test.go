package cmd

import "testing"

const utf8BOM = "\xef\xbb\xbf"

func TestParseImportData(t *testing.T) {
	t.Parallel()

	want := ExportConfig{Name: "prod", Account: "a@b.com", Project: "my-prod-project", Region: "us-central1"}

	yamlBody := "name: prod\naccount: a@b.com\nproject: my-prod-project\nregion: us-central1\n"
	jsonBody := `{"name":"prod","account":"a@b.com","project":"my-prod-project","region":"us-central1"}`

	tests := []struct {
		name string
		data string
		ext  string
	}{
		{name: "yaml", data: yamlBody, ext: ".yaml"},
		{name: "yaml with BOM", data: utf8BOM + yamlBody, ext: ".yaml"},
		{name: "yml with BOM", data: utf8BOM + yamlBody, ext: ".yml"},
		{name: "json", data: jsonBody, ext: ".json"},
		{name: "json with BOM", data: utf8BOM + jsonBody, ext: ".json"},
		{name: "autodetect yaml with BOM", data: utf8BOM + yamlBody, ext: ""},
		{name: "autodetect json with BOM", data: utf8BOM + jsonBody, ext: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseImportData([]byte(tt.data), tt.ext)
			if err != nil {
				t.Fatalf("parseImportData() error = %v", err)
			}
			if got != want {
				t.Errorf("parseImportData() = %+v, want %+v", got, want)
			}
		})
	}
}

func TestParseImportDataInvalid(t *testing.T) {
	t.Parallel()

	if _, err := parseImportData([]byte("name: [unclosed\n"), ".yaml"); err == nil {
		t.Error("parseImportData() error = nil, want an error for malformed YAML")
	}
}
