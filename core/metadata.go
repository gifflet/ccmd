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
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/pkg/errors"
)

const (
	// ConfigFileName is the default name for the configuration file
	ConfigFileName = "ccmd.yaml"
	// LockFileName is the default name for the lock file
	LockFileName = "ccmd-lock.yaml"
)

// LoadProjectConfig loads the project configuration from ccmd.yaml
func LoadProjectConfig(projectPath string) (*ProjectConfig, error) {
	configPath := filepath.Join(projectPath, ConfigFileName)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.FileError("read config", configPath, err)
	}

	var config ProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.FileError("parse config", configPath, err)
	}

	return &config, nil
}

// SaveProjectConfig saves the project configuration to ccmd.yaml
func SaveProjectConfig(projectPath string, config *ProjectConfig) error {
	configPath := filepath.Join(projectPath, ConfigFileName)

	data, err := yaml.Marshal(config)
	if err != nil {
		return errors.FileError("marshal config", configPath, err)
	}

	return os.WriteFile(configPath, data, 0644)
}

// ProjectConfigExists checks if ccmd.yaml exists in the project
func ProjectConfigExists(projectPath string) bool {
	configPath := filepath.Join(projectPath, ConfigFileName)
	_, err := os.Stat(configPath)
	return err == nil
}

// LockFileExists checks if ccmd-lock.yaml exists in the project
func LockFileExists(projectPath string) bool {
	lockPath := filepath.Join(projectPath, LockFileName)
	_, err := os.Stat(lockPath)
	return err == nil
}

// ReadLockFile reads and parses the ccmd-lock.yaml file
func ReadLockFile(path string) (*LockFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.FileError("read lock file", path, err)
	}

	var lock LockFile
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return nil, errors.FileError("parse lock file", path, err)
	}

	if lock.Commands == nil {
		lock.Commands = make(map[string]*LockCommand)
	}

	return &lock, nil
}

// WriteLockFile writes the lock file to disk
func WriteLockFile(path string, lockFile *LockFile) error {
	data, err := yaml.Marshal(lockFile)
	if err != nil {
		return errors.FileError("marshal lock file", path, err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return errors.FileError("write lock file", path, err)
	}

	return nil
}
