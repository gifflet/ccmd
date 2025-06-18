package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/models"
)

// CommandValidator validates command structure and content
type CommandValidator struct {
	commandPath string
}

// NewCommandValidator creates a new validator for a command directory
func NewCommandValidator(commandPath string) *CommandValidator {
	return &CommandValidator{
		commandPath: commandPath,
	}
}

// Validate performs full validation of the command structure
func (v *CommandValidator) Validate() error {
	// Check if path exists
	if _, err := os.Stat(v.commandPath); err != nil {
		return NewValidationError("command directory not found", v.commandPath)
	}

	// Validate ccmd.yaml
	metadata, err := v.validateMetadataFile()
	if err != nil {
		return err
	}

	// Validate index.md
	if err := v.validateIndexFile(); err != nil {
		return err
	}

	// Validate command name matches directory name
	if err := v.validateCommandName(metadata); err != nil {
		return err
	}

	// Validate version format
	if err := v.validateVersion(metadata.Version); err != nil {
		return err
	}

	return nil
}

// validateMetadataFile checks for ccmd.yaml existence and validity
func (v *CommandValidator) validateMetadataFile() (*models.CommandMetadata, error) {
	metadataPath := filepath.Join(v.commandPath, "ccmd.yaml")

	// Check if file exists
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, NewValidationError("ccmd.yaml not found", metadataPath)
		}
		return nil, NewValidationError("failed to read ccmd.yaml", err.Error())
	}

	// Parse YAML
	var metadata models.CommandMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil, NewValidationError("invalid ccmd.yaml format", err.Error())
	}

	// Validate metadata content
	if err := metadata.Validate(); err != nil {
		return nil, NewValidationError("invalid metadata", err.Error())
	}

	return &metadata, nil
}

// validateIndexFile checks for index.md existence
func (v *CommandValidator) validateIndexFile() error {
	indexPath := filepath.Join(v.commandPath, "index.md")

	info, err := os.Stat(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewValidationError("index.md not found", indexPath)
		}
		return NewValidationError("failed to access index.md", err.Error())
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return NewValidationError("index.md is not a regular file", indexPath)
	}

	// Check if file is not empty
	if info.Size() == 0 {
		return NewValidationError("index.md is empty", indexPath)
	}

	return nil
}

// validateCommandName ensures the command name matches the directory name
func (v *CommandValidator) validateCommandName(metadata *models.CommandMetadata) error {
	dirName := filepath.Base(v.commandPath)

	// Handle versioned directory names (e.g., "mycommand@1.0.0")
	parts := strings.Split(dirName, "@")
	expectedName := parts[0]

	if metadata.Name != expectedName {
		return NewValidationError(
			"command name mismatch",
			fmt.Sprintf("expected '%s', got '%s'", expectedName, metadata.Name),
		)
	}

	// If directory has version suffix, validate it matches
	if len(parts) == 2 {
		if metadata.Version != parts[1] {
			return NewValidationError(
				"version mismatch with directory name",
				fmt.Sprintf("expected '%s', got '%s'", parts[1], metadata.Version),
			)
		}
	}

	return nil
}

// validateVersion checks if version follows semantic versioning
func (v *CommandValidator) validateVersion(version string) error {
	// Basic semver pattern (simplified)
	// Matches: 1.0.0, 2.1.0, 0.1.0-beta, 1.0.0-rc.1+build.123
	semverPattern := `^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)` +
		`(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)` +
		`(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?` +
		`(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`

	matched, err := regexp.MatchString(semverPattern, version)
	if err != nil {
		return NewValidationError("version validation failed", err.Error())
	}

	if !matched {
		return NewValidationError(
			"invalid version format",
			fmt.Sprintf("'%s' does not follow semantic versioning", version),
		)
	}

	return nil
}

// ValidateDirectory validates a command directory structure
func ValidateDirectory(path string) error {
	validator := NewCommandValidator(path)
	return validator.Validate()
}

// ValidateInstalled validates an already installed command
func ValidateInstalled(commandsDir, commandName string) error {
	commandPath := filepath.Join(commandsDir, commandName)
	validator := NewCommandValidator(commandPath)
	return validator.Validate()
}

// Validator provides validation methods for commands
type Validator struct {
	fs fs.FileSystem
}

// NewValidator creates a new validator with the given file system
func NewValidator(fileSystem fs.FileSystem) *Validator {
	return &Validator{
		fs: fileSystem,
	}
}

// ValidateCommandStructure validates a command directory structure using the file system
func (v *Validator) ValidateCommandStructure(commandPath string) error {
	// Check if path exists
	exists, err := v.fs.Exists(commandPath)
	if err != nil {
		return NewValidationError("failed to check command directory", err.Error())
	}
	if !exists {
		return NewValidationError("command directory not found", commandPath)
	}

	// Validate ccmd.yaml
	metadataPath := filepath.Join(commandPath, "ccmd.yaml")
	data, err := v.fs.ReadFile(metadataPath)
	if err != nil {
		return NewValidationError("ccmd.yaml not found", metadataPath)
	}

	// Parse YAML
	var metadata models.CommandMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return NewValidationError("invalid ccmd.yaml format", err.Error())
	}

	// Validate metadata content
	if err := metadata.Validate(); err != nil {
		return NewValidationError("invalid metadata", err.Error())
	}

	// Validate index.md
	indexPath := filepath.Join(commandPath, "index.md")
	exists, err = v.fs.Exists(indexPath)
	if err != nil {
		return NewValidationError("failed to check index.md", err.Error())
	}
	if !exists {
		return NewValidationError("index.md not found", indexPath)
	}

	// Check if index.md is not empty
	data, err = v.fs.ReadFile(indexPath)
	if err != nil {
		return NewValidationError("failed to read index.md", err.Error())
	}
	if len(data) == 0 {
		return NewValidationError("index.md is empty", indexPath)
	}

	return nil
}
