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
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gifflet/ccmd/pkg/errors"
	"github.com/gifflet/ccmd/pkg/output"
)

func repoType(cfg *ProjectConfig) string {
	if cfg.Type == "plugin" {
		return "plugin"
	}
	return "command"
}

// installPlugin copies a cloned plugin repo into .claude/plugins/{name} and
// registers it in .claude/settings.json and ccmd-lock.yaml.
func installPlugin(projectRoot, tempDir string, cfg *ProjectConfig, opts InstallOptions) (string, error) {
	name := opts.Name
	if name == "" {
		name = cfg.Name
	}
	if name == "" {
		name = extractCommandName(opts.Repository)
	}

	if err := validateCommandName(name); err != nil {
		return "", err
	}

	pluginsDir := filepath.Join(projectRoot, ".claude", "plugins")
	if err := os.MkdirAll(pluginsDir, 0o750); err != nil {
		return "", errors.FileError("create plugins directory", pluginsDir, err)
	}

	targetRepoPath := ExtractRepoPath(opts.Repository)
	existingPlugin, err := findExistingPluginByRepo(projectRoot, targetRepoPath)
	if err != nil {
		return "", errors.FileError("check existing plugins", "", err)
	}

	if existingPlugin != "" && !opts.Force {
		return "", errors.AlreadyExists(fmt.Sprintf(
			"repository already installed as plugin %q, use --force to reinstall",
			existingPlugin))
	}

	if opts.Force && existingPlugin != "" {
		output.PrintInfof("Removing previous installation %q...", existingPlugin)
		if err := removePlugin(projectRoot, existingPlugin); err != nil {
			return "", err
		}
	}

	destDir := filepath.Join(pluginsDir, name)
	output.PrintInfof("Installing plugin %q...", name)
	if err := copyDirectory(tempDir, destDir); err != nil {
		return "", errors.FileError("copy plugin files", destDir, err)
	}

	originalVersion := cfg.Version
	cfg.Name = name
	cfg.Repository = opts.Repository

	if err := writeCommandMetadata(filepath.Join(destDir, "ccmd.yaml"), cfg); err != nil {
		if removeErr := os.RemoveAll(destDir); removeErr != nil {
			output.PrintWarningf("Failed to cleanup plugin directory: %v", removeErr)
		}
		return "", err
	}

	if err := enablePlugin(projectRoot, name); err != nil {
		output.PrintWarningf("Failed to register plugin in settings.json: %v", err)
	}

	if err := updatePluginLockFile(projectRoot, name, cfg, originalVersion, opts.Version); err != nil {
		output.PrintWarningf("Failed to update lock file: %v", err)
	}

	repoSpec := opts.Repository
	if strings.Contains(repoSpec, "://") || strings.HasPrefix(repoSpec, "git@") {
		repoSpec = ExtractRepoPath(repoSpec)
	}
	versionForConfig := opts.Version
	if isCommitHash(versionForConfig) && len(versionForConfig) > 7 {
		versionForConfig = versionForConfig[:7]
	}
	if err := addPluginToConfig(projectRoot, name, repoSpec, versionForConfig); err != nil {
		output.PrintWarningf("Failed to update ccmd.yaml: %v", err)
	}

	output.PrintSuccessf("Plugin %q installed successfully", name)
	printPluginComponents(scanPluginComponents(destDir))
	return name, nil
}

// removePlugin deletes a plugin installation and removes it from settings and lock file.
func removePlugin(projectRoot, name string) error {
	pluginDir := filepath.Join(projectRoot, ".claude", "plugins", name)

	if dirExists(pluginDir) {
		output.PrintInfof("Removing plugin directory...")
		if err := os.RemoveAll(pluginDir); err != nil {
			return errors.FileError("remove plugin directory", pluginDir, err)
		}
	}

	if err := disablePlugin(projectRoot, name); err != nil {
		output.PrintWarningf("Failed to remove plugin from settings.json: %v", err)
	}

	lockPath := filepath.Join(projectRoot, LockFileName)
	if fileExists(lockPath) {
		lockFile, err := ReadLockFile(lockPath)
		if err == nil {
			delete(lockFile.Plugins, name)
			if writeErr := WriteLockFile(lockPath, lockFile); writeErr != nil {
				output.PrintWarningf("Failed to update lock file: %v", writeErr)
			}
		}
	}

	return nil
}

func findExistingPluginByRepo(projectRoot, targetRepoPath string) (string, error) {
	pluginsDir := filepath.Join(projectRoot, ".claude", "plugins")
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		metadataPath := filepath.Join(pluginsDir, entry.Name(), "ccmd.yaml")
		if metadata, err := readCommandMetadata(metadataPath); err == nil && metadata.Repository != "" {
			if ExtractRepoPath(metadata.Repository) == targetRepoPath {
				return entry.Name(), nil
			}
		}
	}

	return "", nil
}

func addPluginToConfig(projectRoot, _, repository, version string) error {
	var config *ProjectConfig
	if ProjectConfigExists(projectRoot) {
		var err error
		config, err = LoadProjectConfig(projectRoot)
		if err != nil {
			return err
		}
	} else {
		config = &ProjectConfig{}
	}

	pluginSpec := repository
	if version != "" {
		pluginSpec = fmt.Sprintf("%s@%s", repository, version)
	}

	currentRepo := ExtractRepoPath(repository)
	found := false

	for i, spec := range config.Plugins {
		repo, _ := ParseCommandSpec(spec)
		if ExtractRepoPath(repo) == currentRepo {
			config.Plugins[i] = pluginSpec
			found = true
			break
		}
	}

	if !found {
		config.Plugins = append(config.Plugins, pluginSpec)
	}

	return SaveProjectConfig(projectRoot, config)
}

func removePluginFromConfig(projectRoot, name, repository string) error {
	configPath := filepath.Join(projectRoot, ConfigFileName)
	if !fileExists(configPath) {
		return nil
	}

	config, err := LoadProjectConfig(projectRoot)
	if err != nil {
		return err
	}

	currentRepo := ExtractRepoPath(repository)
	newPlugins := make([]string, 0, len(config.Plugins))
	for _, spec := range config.Plugins {
		repo, _ := ParseCommandSpec(spec)
		if ExtractRepoPath(repo) == currentRepo || extractCommandName(repo) == name {
			continue
		}
		newPlugins = append(newPlugins, spec)
	}

	config.Plugins = newPlugins
	return SaveProjectConfig(projectRoot, config)
}

func updatePluginLockFile(
	projectRoot, name string,
	cfg *ProjectConfig,
	originalVersion, requestedVersion string,
) error {
	lockPath := filepath.Join(projectRoot, LockFileName)
	now := time.Now()

	var lockFile *LockFile
	if fileExists(lockPath) {
		var err error
		lockFile, err = ReadLockFile(lockPath)
		if err != nil {
			return err
		}
	} else {
		lockFile = &LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands:        make(map[string]*LockCommand),
			Plugins:         make(map[string]*LockPlugin),
		}
	}

	commitHash := "unknown"
	pluginPath := filepath.Join(projectRoot, ".claude", "plugins", name)
	if hash, err := gitGetCurrentCommit(pluginPath); err == nil {
		commitHash = hash
	}

	resolved := cfg.Repository
	if requestedVersion != "" {
		resolved = fmt.Sprintf("%s@%s", cfg.Repository, requestedVersion)
	} else {
		if defaultBranch, err := gitGetDefaultBranch(pluginPath); err == nil {
			resolved = fmt.Sprintf("%s@%s", cfg.Repository, defaultBranch)
		} else if commitHash != "unknown" && len(commitHash) >= 7 {
			resolved = fmt.Sprintf("%s@%s", cfg.Repository, commitHash[:7])
		}
	}

	repoPath := ExtractRepoPath(cfg.Repository)
	var existingKey string
	var existingPlugin *LockPlugin

	for key, p := range lockFile.Plugins {
		if ExtractRepoPath(p.Source) == repoPath {
			existingKey = key
			existingPlugin = p
			break
		}
	}

	installedAt := now
	if existingPlugin != nil && !existingPlugin.InstalledAt.IsZero() {
		installedAt = existingPlugin.InstalledAt
	}

	if existingKey != "" && existingKey != name {
		delete(lockFile.Plugins, existingKey)
	}

	lockFile.Plugins[name] = &LockPlugin{
		Name:        name,
		Version:     originalVersion,
		Source:      cfg.Repository,
		Resolved:    resolved,
		Commit:      commitHash,
		InstalledAt: installedAt,
		UpdatedAt:   now,
	}

	return WriteLockFile(lockPath, lockFile)
}

type ccmdMarketplace struct {
	Name    string                  `json:"name"`
	Owner   map[string]string       `json:"owner"`
	Plugins []ccmdMarketplacePlugin `json:"plugins"`
}

type ccmdMarketplacePlugin struct {
	Name        string `json:"name"`
	Source      string `json:"source"`
	Description string `json:"description,omitempty"`
}

// enablePlugin adds the plugin to .claude/settings.json enabledPlugins and
// registers the ccmd marketplace in extraKnownMarketplaces.
func enablePlugin(projectRoot, name string) error {
	claudeDir := filepath.Join(projectRoot, ".claude")
	settings, err := ReadClaudeSettings(claudeDir)
	if err != nil {
		return err
	}

	if settings.EnabledPlugins == nil {
		settings.EnabledPlugins = make(map[string]bool)
	}
	settings.EnabledPlugins[fmt.Sprintf("%s@ccmd", name)] = true

	pluginsDir := filepath.Join(projectRoot, ".claude", "plugins")
	absPluginsDir, absErr := filepath.Abs(pluginsDir)
	if absErr != nil {
		absPluginsDir = pluginsDir
	}

	if settings.ExtraKnownMarketplaces == nil {
		settings.ExtraKnownMarketplaces = make(map[string]MarketplaceEntry)
	}
	settings.ExtraKnownMarketplaces["ccmd"] = MarketplaceEntry{
		Source: MarketplaceSource{
			Source: "directory",
			Path:   absPluginsDir,
		},
	}

	if err := WriteClaudeSettings(claudeDir, settings); err != nil {
		return err
	}

	return updateCCMDMarketplace(projectRoot, name, true)
}

// disablePlugin removes the plugin from .claude/settings.json enabledPlugins.
func disablePlugin(projectRoot, name string) error {
	claudeDir := filepath.Join(projectRoot, ".claude")
	settings, err := ReadClaudeSettings(claudeDir)
	if err != nil {
		return err
	}

	delete(settings.EnabledPlugins, fmt.Sprintf("%s@ccmd", name))

	if err := WriteClaudeSettings(claudeDir, settings); err != nil {
		return err
	}

	return updateCCMDMarketplace(projectRoot, name, false)
}

type pluginComponents struct {
	Commands   []string
	Skills     []string
	Agents     []string
	MCPServers []string
}

func scanMDFilesInDir(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") && strings.HasSuffix(e.Name(), ".md") {
			names = append(names, strings.TrimSuffix(e.Name(), ".md"))
		}
	}
	return names
}

func scanSkillsDir(pluginDir string) []string {
	entries, err := os.ReadDir(filepath.Join(pluginDir, "skills"))
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		skillFile := filepath.Join(pluginDir, "skills", e.Name(), "SKILL.md")
		if _, err := os.Stat(skillFile); err == nil {
			names = append(names, e.Name())
		}
	}
	return names
}

func scanMCPServers(pluginDir string) []string {
	data, err := os.ReadFile(filepath.Join(pluginDir, ".mcp.json"))
	if err != nil {
		return nil
	}
	var mcpConfig struct {
		MCPServers map[string]json.RawMessage `json:"mcpServers"`
	}
	if json.Unmarshal(data, &mcpConfig) != nil {
		return nil
	}
	names := make([]string, 0, len(mcpConfig.MCPServers))
	for name := range mcpConfig.MCPServers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func scanPluginComponents(pluginDir string) pluginComponents {
	return pluginComponents{
		Commands:   scanMDFilesInDir(filepath.Join(pluginDir, "commands")),
		Skills:     scanSkillsDir(pluginDir),
		Agents:     scanMDFilesInDir(filepath.Join(pluginDir, "agents")),
		MCPServers: scanMCPServers(pluginDir),
	}
}

func printPluginComponents(c pluginComponents) {
	if len(c.Agents)+len(c.Commands)+len(c.Skills)+len(c.MCPServers) == 0 {
		return
	}
	output.PrintInfof("Components installed:")
	if len(c.Agents) > 0 {
		output.PrintInfof("  Agents:      %s", strings.Join(c.Agents, ", "))
	}
	if len(c.Commands) > 0 {
		output.PrintInfof("  Commands:    /%s", strings.Join(c.Commands, ", /"))
	}
	if len(c.Skills) > 0 {
		output.PrintInfof("  Skills:      %s", strings.Join(c.Skills, ", "))
	}
	if len(c.MCPServers) > 0 {
		output.PrintInfof("  MCP servers: %s", strings.Join(c.MCPServers, ", "))
	}
}

// updateCCMDMarketplace maintains .claude/plugins/.claude-plugin/marketplace.json.
func updateCCMDMarketplace(projectRoot, pluginName string, add bool) error {
	pluginsDir := filepath.Join(projectRoot, ".claude", "plugins")
	marketplaceDir := filepath.Join(pluginsDir, ".claude-plugin")
	if err := os.MkdirAll(marketplaceDir, 0o750); err != nil {
		return errors.FileError("create marketplace directory", marketplaceDir, err)
	}

	marketplacePath := filepath.Join(marketplaceDir, "marketplace.json")

	var marketplace ccmdMarketplace
	if data, err := os.ReadFile(marketplacePath); err == nil {
		_ = json.Unmarshal(data, &marketplace)
	}

	if marketplace.Name == "" {
		marketplace.Name = "ccmd"
		marketplace.Owner = map[string]string{"name": "ccmd"}
	}

	updated := make([]ccmdMarketplacePlugin, 0, len(marketplace.Plugins))
	for _, p := range marketplace.Plugins {
		if p.Name != pluginName {
			updated = append(updated, p)
		}
	}

	if add {
		description := ""
		metadataPath := filepath.Join(pluginsDir, pluginName, "ccmd.yaml")
		if metadata, err := readCommandMetadata(metadataPath); err == nil {
			description = metadata.Description
		}
		updated = append(updated, ccmdMarketplacePlugin{
			Name:        pluginName,
			Source:      "./" + pluginName,
			Description: description,
		})
	}

	marketplace.Plugins = updated

	data, err := json.MarshalIndent(marketplace, "", "  ")
	if err != nil {
		return errors.FileError("marshal marketplace", marketplacePath, err)
	}

	return os.WriteFile(marketplacePath, data, 0o600)
}
