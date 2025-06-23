// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package model

// RepositoryConfig holds configuration for different repository types
type RepositoryConfig struct {
	Type   string                 `yaml:"type" json:"type"`
	Config map[string]interface{} `yaml:"config" json:"config"`
}

// CommandInfo provides basic metadata about a command
type CommandInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Repository  string `json:"repository,omitempty"`
}

// RepositoryInfo contains metadata about a repository
// TODO: Implement full structure with status, capabilities, and statistics
type RepositoryInfo struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CacheInfo contains information about repository cache
// TODO: Implement full structure with cache statistics and metadata
type CacheInfo struct {
	Enabled     bool    `json:"enabled"`
	Size        int64   `json:"size"`
	Entries     int     `json:"entries"`
	LastUpdated string  `json:"last_updated"`
	TTL         int     `json:"ttl_seconds"`
	HitRate     float64 `json:"hit_rate"`
}
