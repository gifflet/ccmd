/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/internal/fs"
	ccmderrors "github.com/gifflet/ccmd/pkg/errors"
)

// CommandInfo represents the structured information about a command
type CommandInfo struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Author      string            `json:"author"`
	Description string            `json:"description"`
	Repository  string            `json:"repository"`
	License     string            `json:"license,omitempty"`
	Homepage    string            `json:"homepage,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Entry       string            `json:"entry,omitempty"`
	Source      string            `json:"source"`
	InstalledAt string            `json:"installed_at"`
	UpdatedAt   string            `json:"updated_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Structure   StructureInfo     `json:"structure"`
}

// StructureInfo contains information about command structure integrity
type StructureInfo struct {
	DirectoryExists bool     `json:"directory_exists"`
	MarkdownExists  bool     `json:"markdown_exists"`
	HasCcmdYaml     bool     `json:"has_ccmd_yaml"`
	HasIndexMd      bool     `json:"has_index_md"`
	IsValid         bool     `json:"is_valid"`
	Issues          []string `json:"issues,omitempty"`
}

// GetCommandDetails retrieves detailed information about an installed command
func GetCommandDetails(commandName, projectPath string, filesystem fs.FileSystem) (*CommandInfo, error) {
	if filesystem == nil {
		filesystem = fs.OS{}
	}

	// Get basic command info from lock file
	lockInfo, err := GetCommandInfo(commandName, projectPath)
	if err != nil {
		if errors.Is(err, ccmderrors.ErrNotFound) {
			return nil, fmt.Errorf("command '%s' is not installed", commandName)
		}
		return nil, fmt.Errorf("failed to get command info: %w", err)
	}

	if lockInfo == nil {
		return nil, fmt.Errorf("command '%s' is not installed", commandName)
	}

	// Get base directory
	baseDir := filepath.Join(projectPath, ".claude")

	// Check structure and get metadata
	structureInfo, metadata := checkCommandStructure(commandName, baseDir, filesystem)

	// Build command info
	info := &CommandInfo{
		Name:        lockInfo.Name,
		Version:     lockInfo.Version,
		Source:      lockInfo.Repository,
		Repository:  lockInfo.Repository,
		InstalledAt: lockInfo.InstalledAt,
		UpdatedAt:   lockInfo.UpdatedAt,
		Metadata:    make(map[string]string),
		Structure:   structureInfo,
	}

	// Fill in data from ccmd.yaml if available
	if metadata != nil {
		info.Author = metadata.Author
		info.Description = metadata.Description
		info.Repository = metadata.Repository
		info.License = metadata.License
		info.Homepage = metadata.Homepage
		info.Tags = metadata.Tags
		info.Entry = metadata.Entry
	} else if lockInfo.Description != "" {
		// Fallback to lock file metadata
		info.Description = lockInfo.Description
	}

	return info, nil
}

// FormatCommandInfoJSON formats command info as indented JSON
func FormatCommandInfoJSON(info *CommandInfo) (string, error) {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}

// ReadCommandContentPreview reads a preview of the command's index.md file
func ReadCommandContentPreview(commandName, baseDir string, filesystem fs.FileSystem, lines int) (string, int, error) {
	indexPath := filepath.Join(baseDir, "commands", commandName, "index.md")

	content, err := filesystem.ReadFile(indexPath)
	if err != nil {
		return "", 0, err
	}

	contentStr := string(content)
	allLines := splitLines(contentStr)

	previewLines := lines
	if len(allLines) < previewLines {
		previewLines = len(allLines)
	}

	preview := ""
	for i := 0; i < previewLines; i++ {
		preview += allLines[i] + "\n"
	}

	return preview, len(allLines), nil
}

func checkCommandStructure(commandName, baseDir string, filesystem fs.FileSystem) (StructureInfo, *ProjectConfig) {
	info := StructureInfo{
		DirectoryExists: false,
		MarkdownExists:  false,
		HasCcmdYaml:     false,
		HasIndexMd:      false,
		IsValid:         false,
		Issues:          []string{},
	}

	commandDir := filepath.Join(baseDir, "commands", commandName)
	markdownFile := filepath.Join(baseDir, "commands", commandName+".md")
	ccmdYamlFile := filepath.Join(commandDir, "ccmd.yaml")
	indexMdFile := filepath.Join(commandDir, "index.md")

	// Check directory
	if dirInfo, err := filesystem.Stat(commandDir); err == nil && dirInfo.IsDir() {
		info.DirectoryExists = true
	} else {
		info.Issues = append(info.Issues, "Command directory is missing")
	}

	// Check markdown file
	if fileInfo, err := filesystem.Stat(markdownFile); err == nil && !fileInfo.IsDir() {
		info.MarkdownExists = true
	} else {
		info.Issues = append(info.Issues, "Standalone markdown file is missing")
	}

	// Check ccmd.yaml
	var metadata *ProjectConfig
	if info.DirectoryExists {
		if data, err := filesystem.ReadFile(ccmdYamlFile); err == nil {
			info.HasCcmdYaml = true
			metadata = &ProjectConfig{}
			if err := yaml.Unmarshal(data, metadata); err != nil {
				info.Issues = append(info.Issues, "ccmd.yaml is malformed")
				metadata = nil
			}
		} else {
			info.Issues = append(info.Issues, "ccmd.yaml is missing")
		}

		// Check index.md
		if _, err := filesystem.Stat(indexMdFile); err == nil {
			info.HasIndexMd = true
		}
	}

	// Determine if structure is valid
	info.IsValid = info.DirectoryExists && info.MarkdownExists && info.HasCcmdYaml
	if !info.IsValid && len(info.Issues) == 0 {
		info.Issues = append(info.Issues, "Incomplete command structure")
	}

	return info, metadata
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
