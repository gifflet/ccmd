package models

import (
	"reflect"
	"testing"
)

func TestCommandMetadata_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cm      CommandMetadata
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid metadata",
			cm: CommandMetadata{
				Name:        "test-command",
				Version:     "1.0.0",
				Description: "A test command",
				Author:      "Test Author",
				Repository:  "github.com/test/repo",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			cm: CommandMetadata{
				Version:     "1.0.0",
				Description: "A test command",
				Author:      "Test Author",
				Repository:  "github.com/test/repo",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing version",
			cm: CommandMetadata{
				Name:        "test-command",
				Description: "A test command",
				Author:      "Test Author",
				Repository:  "github.com/test/repo",
			},
			wantErr: true,
			errMsg:  "version is required",
		},
		{
			name: "missing description",
			cm: CommandMetadata{
				Name:       "test-command",
				Version:    "1.0.0",
				Author:     "Test Author",
				Repository: "github.com/test/repo",
			},
			wantErr: true,
			errMsg:  "description is required",
		},
		{
			name: "missing author",
			cm: CommandMetadata{
				Name:        "test-command",
				Version:     "1.0.0",
				Description: "A test command",
				Repository:  "github.com/test/repo",
			},
			wantErr: true,
			errMsg:  "author is required",
		},
		{
			name: "missing repository",
			cm: CommandMetadata{
				Name:        "test-command",
				Version:     "1.0.0",
				Description: "A test command",
				Author:      "Test Author",
			},
			wantErr: true,
			errMsg:  "repository is required",
		},
		{
			name: "full metadata with optional fields",
			cm: CommandMetadata{
				Name:        "test-command",
				Version:     "1.0.0",
				Description: "A test command",
				Author:      "Test Author",
				Repository:  "github.com/test/repo",
				Entry:       "main.go",
				Tags:        []string{"cli", "tool"},
				License:     "MIT",
				Homepage:    "https://example.com",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cm.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestCommandMetadata_YAML(t *testing.T) {
	cm := CommandMetadata{
		Name:        "test-command",
		Version:     "1.0.0",
		Description: "A test command",
		Author:      "Test Author",
		Repository:  "github.com/test/repo",
		Entry:       "main.go",
		Tags:        []string{"cli", "tool"},
		License:     "MIT",
		Homepage:    "https://example.com",
	}

	// Test MarshalYAML
	data, err := cm.MarshalYAML()
	if err != nil {
		t.Fatalf("MarshalYAML() error = %v", err)
	}

	// Test UnmarshalYAML
	var cm2 CommandMetadata
	err = cm2.UnmarshalYAML(data)
	if err != nil {
		t.Fatalf("UnmarshalYAML() error = %v", err)
	}

	if !reflect.DeepEqual(cm, cm2) {
		t.Errorf("YAML round-trip failed: got %+v, want %+v", cm2, cm)
	}
}

func TestCommandMetadata_JSON(t *testing.T) {
	cm := CommandMetadata{
		Name:        "test-command",
		Version:     "1.0.0",
		Description: "A test command",
		Author:      "Test Author",
		Repository:  "github.com/test/repo",
		Entry:       "main.go",
		Tags:        []string{"cli", "tool"},
		License:     "MIT",
		Homepage:    "https://example.com",
	}

	// Test MarshalJSON
	data, err := cm.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	// Test UnmarshalJSON
	var cm2 CommandMetadata
	err = cm2.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	if !reflect.DeepEqual(cm, cm2) {
		t.Errorf("JSON round-trip failed: got %+v, want %+v", cm2, cm)
	}
}

func TestCommandMetadata_OptionalFields(t *testing.T) {
	// Test that optional fields are properly omitted when empty
	cm := CommandMetadata{
		Name:        "test-command",
		Version:     "1.0.0",
		Description: "A test command",
		Author:      "Test Author",
		Repository:  "github.com/test/repo",
	}

	yamlData, err := cm.MarshalYAML()
	if err != nil {
		t.Fatalf("MarshalYAML() error = %v", err)
	}

	yamlStr := string(yamlData)

	// Check that optional fields are not in the YAML output
	if contains(yamlStr, "entry:") {
		t.Error("YAML should not contain empty entry field")
	}
	if contains(yamlStr, "tags:") {
		t.Error("YAML should not contain empty tags field")
	}
	if contains(yamlStr, "license:") {
		t.Error("YAML should not contain empty license field")
	}
	if contains(yamlStr, "homepage:") {
		t.Error("YAML should not contain empty homepage field")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
