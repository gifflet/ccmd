/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gifflet/ccmd/pkg/commands"
	"github.com/gifflet/ccmd/pkg/project"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	assert.Equal(t, "sync", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check flags
	dryRunFlag := cmd.Flags().Lookup("dry-run")
	assert.NotNil(t, dryRunFlag)
	assert.Equal(t, "false", dryRunFlag.DefValue)

	forceFlag := cmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "false", forceFlag.DefValue)
	assert.Equal(t, "f", forceFlag.Shorthand)
}

func TestAnalyzeSync(t *testing.T) {
	tests := []struct {
		name           string
		configCommands map[string]project.ConfigCommand
		installedMap   map[string]*commands.CommandDetail
		wantToInstall  []string
		wantToRemove   []string
	}{
		{
			name: "all in sync",
			configCommands: map[string]project.ConfigCommand{
				"tool1": {Repo: "owner/tool1", Version: "v1.0.0"},
				"tool2": {Repo: "owner/tool2", Version: "v2.0.0"},
			},
			installedMap: map[string]*commands.CommandDetail{
				"tool1": {CommandLockInfo: &project.CommandLockInfo{Name: "tool1"}},
				"tool2": {CommandLockInfo: &project.CommandLockInfo{Name: "tool2"}},
			},
			wantToInstall: []string{},
			wantToRemove:  []string{},
		},
		{
			name: "need to install",
			configCommands: map[string]project.ConfigCommand{
				"tool1": {Repo: "owner/tool1", Version: "v1.0.0"},
				"tool2": {Repo: "owner/tool2", Version: "v2.0.0"},
			},
			installedMap: map[string]*commands.CommandDetail{
				"tool1": {CommandLockInfo: &project.CommandLockInfo{Name: "tool1"}},
			},
			wantToInstall: []string{"tool2"},
			wantToRemove:  []string{},
		},
		{
			name: "need to remove",
			configCommands: map[string]project.ConfigCommand{
				"tool1": {Repo: "owner/tool1", Version: "v1.0.0"},
			},
			installedMap: map[string]*commands.CommandDetail{
				"tool1": {CommandLockInfo: &project.CommandLockInfo{Name: "tool1"}},
				"tool2": {CommandLockInfo: &project.CommandLockInfo{Name: "tool2"}},
			},
			wantToInstall: []string{},
			wantToRemove:  []string{"tool2"},
		},
		{
			name: "need both install and remove",
			configCommands: map[string]project.ConfigCommand{
				"tool1": {Repo: "owner/tool1", Version: "v1.0.0"},
				"tool3": {Repo: "owner/tool3", Version: "v3.0.0"},
			},
			installedMap: map[string]*commands.CommandDetail{
				"tool1": {CommandLockInfo: &project.CommandLockInfo{Name: "tool1"}},
				"tool2": {CommandLockInfo: &project.CommandLockInfo{Name: "tool2"}},
			},
			wantToInstall: []string{"tool3"},
			wantToRemove:  []string{"tool2"},
		},
		{
			name:           "empty config",
			configCommands: map[string]project.ConfigCommand{},
			installedMap: map[string]*commands.CommandDetail{
				"tool1": {CommandLockInfo: &project.CommandLockInfo{Name: "tool1"}},
				"tool2": {CommandLockInfo: &project.CommandLockInfo{Name: "tool2"}},
			},
			wantToInstall: []string{},
			wantToRemove:  []string{"tool1", "tool2"},
		},
		{
			name: "empty installed",
			configCommands: map[string]project.ConfigCommand{
				"tool1": {Repo: "owner/tool1", Version: "v1.0.0"},
				"tool2": {Repo: "owner/tool2", Version: "v2.0.0"},
			},
			installedMap:  map[string]*commands.CommandDetail{},
			wantToInstall: []string{"tool1", "tool2"},
			wantToRemove:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzeSync(tt.configCommands, tt.installedMap)

			// Sort for consistent comparison
			assert.ElementsMatch(t, tt.wantToInstall, result.ToInstall)
			assert.ElementsMatch(t, tt.wantToRemove, result.ToRemove)
			assert.Empty(t, result.Errors)
		})
	}
}

func TestIsConfirmation(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"y", true},
		{"Y", true},
		{"yes", true},
		{"YES", true},
		{"Yes", true},
		{" y ", true},
		{" yes ", true},
		{"n", false},
		{"N", false},
		{"no", false},
		{"NO", false},
		{"", false},
		{"maybe", false},
		{"yep", false},
		{"yeah", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isConfirmation(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCountErrors(t *testing.T) {
	errors := []error{
		assert.AnError,
		assert.AnError,
		assert.AnError,
	}

	// Mock error messages
	errors[0] = &mockError{msg: "failed to install tool1: connection error"}
	errors[1] = &mockError{msg: "failed to remove tool2: permission denied"}
	errors[2] = &mockError{msg: "failed to install tool3: timeout"}

	count := countErrors(errors, "install")
	assert.Equal(t, 2, count)

	count = countErrors(errors, "remove")
	assert.Equal(t, 1, count)

	count = countErrors(errors, "update")
	assert.Equal(t, 0, count)
}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

func TestSyncResult(t *testing.T) {
	result := Result{
		ToInstall: []string{"tool1", "tool2"},
		ToRemove:  []string{"tool3"},
		Errors:    []error{},
	}

	assert.Len(t, result.ToInstall, 2)
	assert.Len(t, result.ToRemove, 1)
	assert.Empty(t, result.Errors)

	// Add error
	result.Errors = append(result.Errors, assert.AnError)
	assert.Len(t, result.Errors, 1)
}

// TestUpdateLockFile tests lock file updates
func TestUpdateLockFile(t *testing.T) {
	// This test would require mocking the project manager and file system
	// For now, we'll just ensure the function exists and compiles
	t.Skip("Requires filesystem mocking")
}
