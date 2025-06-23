/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package repository

import (
	"fmt"
	"sync"
)

// manager implements the Manager interface for handling repository factories
type manager struct {
	mu        sync.RWMutex
	factories map[string]Factory
}

// NewManager creates a new repository manager instance
func NewManager() Manager {
	return &manager{
		factories: make(map[string]Factory),
	}
}

// Register adds a new repository type with its factory function
func (m *manager) Register(repoType string, factory Factory) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if factory == nil {
		panic(fmt.Sprintf("repository: Register factory for type %s is nil", repoType))
	}

	if _, exists := m.factories[repoType]; exists {
		panic(fmt.Sprintf("repository: Register called twice for type %s", repoType))
	}

	m.factories[repoType] = factory
}

// Create instantiates a new repository of the specified type
func (m *manager) Create(repoType string, config map[string]interface{}) (Repository, error) {
	m.mu.RLock()
	factory, exists := m.factories[repoType]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("repository type %q not registered", repoType)
	}

	repo, err := factory(config)
	if err != nil {
		return nil, NewRepositoryError("create", repoType, err, "factory creation failed")
	}

	return repo, nil
}

// List returns all registered repository types
func (m *manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	types := make([]string, 0, len(m.factories))
	for repoType := range m.factories {
		types = append(types, repoType)
	}

	return types
}

// DefaultManager is the global repository manager instance
var DefaultManager = NewManager()

// Register adds a repository type to the default manager
func Register(repoType string, factory Factory) {
	DefaultManager.Register(repoType, factory)
}

// Create creates a repository using the default manager
func Create(repoType string, config map[string]interface{}) (Repository, error) {
	return DefaultManager.Create(repoType, config)
}

// List returns all registered types from the default manager
func List() []string {
	return DefaultManager.List()
}
