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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestLockFile_Marshal(t *testing.T) {
	t.Run("marshal empty lock file", func(t *testing.T) {
		lockFile := LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands:        make(map[string]*LockCommand),
		}

		data, err := yaml.Marshal(&lockFile)
		require.NoError(t, err)

		expected := `version: "1.0"
lockfileVersion: 1
commands: {}
`
		assert.Equal(t, expected, string(data))
	})

	t.Run("marshal lock file with commands", func(t *testing.T) {
		now := time.Date(2025, 1, 6, 12, 0, 0, 0, time.UTC)
		lockFile := LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands: map[string]*LockCommand{
				"hello-world": {
					Name:        "hello-world",
					Version:     "1.0.0",
					Source:      "https://github.com/gifflet/hello-world.git",
					Resolved:    "https://github.com/gifflet/hello-world.git@v1.0.0",
					Commit:      "05d746d17f6e2235ad9a93acc307b68caa18a281",
					InstalledAt: now,
					UpdatedAt:   now,
				},
			},
		}

		data, err := yaml.Marshal(&lockFile)
		require.NoError(t, err)
		assert.Contains(t, string(data), "version: \"1.0\"")
		assert.Contains(t, string(data), "lockfileVersion: 1")
		assert.Contains(t, string(data), "hello-world:")
		assert.Contains(t, string(data), "name: hello-world")
		assert.Contains(t, string(data), "version: 1.0.0")
		assert.Contains(t, string(data), "source: https://github.com/gifflet/hello-world.git")
		assert.Contains(t, string(data), "resolved: https://github.com/gifflet/hello-world.git@v1.0.0")
		assert.Contains(t, string(data), "commit: 05d746d17f6e2235ad9a93acc307b68caa18a281")
	})
}

func TestLockFile_Unmarshal(t *testing.T) {
	t.Run("unmarshal valid lock file", func(t *testing.T) {
		yamlData := `version: "1.0"
lockfileVersion: 1
commands:
    hello-world:
        name: hello-world
        version: 1.0.0
        source: https://github.com/gifflet/hello-world.git
        resolved: https://github.com/gifflet/hello-world.git@v1.0.0
        commit: 05d746d17f6e2235ad9a93acc307b68caa18a281
        installed_at: 2025-01-06T12:00:00Z
        updated_at: 2025-01-06T12:00:00Z
    parallax:
        name: parallax
        version: 1.0.0
        source: https://github.com/gifflet/parallax.git
        resolved: https://github.com/gifflet/parallax.git@8f3a30e
        commit: 8f3a30ee0340cc51d584c215a6635d445b4a7c56
        installed_at: 2025-01-06T13:00:00Z
        updated_at: 2025-01-06T13:00:00Z
`

		var lockFile LockFile
		err := yaml.Unmarshal([]byte(yamlData), &lockFile)
		require.NoError(t, err)

		assert.Equal(t, "1.0", lockFile.Version)
		assert.Equal(t, 1, lockFile.LockfileVersion)
		assert.Len(t, lockFile.Commands, 2)

		// Check hello-world command
		helloWorld := lockFile.Commands["hello-world"]
		require.NotNil(t, helloWorld)
		assert.Equal(t, "hello-world", helloWorld.Name)
		assert.Equal(t, "1.0.0", helloWorld.Version)
		assert.Equal(t, "https://github.com/gifflet/hello-world.git", helloWorld.Source)
		assert.Equal(t, "https://github.com/gifflet/hello-world.git@v1.0.0", helloWorld.Resolved)
		assert.Equal(t, "05d746d17f6e2235ad9a93acc307b68caa18a281", helloWorld.Commit)
		assert.False(t, helloWorld.InstalledAt.IsZero())
		assert.False(t, helloWorld.UpdatedAt.IsZero())

		// Check parallax command
		parallax := lockFile.Commands["parallax"]
		require.NotNil(t, parallax)
		assert.Equal(t, "parallax", parallax.Name)
		assert.Equal(t, "1.0.0", parallax.Version)
		assert.Equal(t, "https://github.com/gifflet/parallax.git", parallax.Source)
		assert.Equal(t, "https://github.com/gifflet/parallax.git@8f3a30e", parallax.Resolved)
		assert.Equal(t, "8f3a30ee0340cc51d584c215a6635d445b4a7c56", parallax.Commit)
	})

	t.Run("unmarshal empty commands", func(t *testing.T) {
		yamlData := `version: "1.0"
lockfileVersion: 1
commands: {}
`

		var lockFile LockFile
		err := yaml.Unmarshal([]byte(yamlData), &lockFile)
		require.NoError(t, err)

		assert.Equal(t, "1.0", lockFile.Version)
		assert.Equal(t, 1, lockFile.LockfileVersion)
		assert.NotNil(t, lockFile.Commands)
		assert.Len(t, lockFile.Commands, 0)
	})
}

func TestLockFile_RoundTrip(t *testing.T) {
	t.Run("marshal and unmarshal preserves data", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)
		original := LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands: map[string]*LockCommand{
				"test-cmd": {
					Name:        "test-cmd",
					Version:     "2.1.0",
					Source:      "https://github.com/user/test-cmd.git",
					Resolved:    "https://github.com/user/test-cmd.git@v2.1.0",
					Commit:      "abc123def456",
					InstalledAt: now,
					UpdatedAt:   now.Add(time.Hour),
				},
			},
		}

		// Marshal
		data, err := yaml.Marshal(&original)
		require.NoError(t, err)

		// Unmarshal
		var result LockFile
		err = yaml.Unmarshal(data, &result)
		require.NoError(t, err)

		// Compare
		assert.Equal(t, original.Version, result.Version)
		assert.Equal(t, original.LockfileVersion, result.LockfileVersion)
		assert.Len(t, result.Commands, 1)

		cmd := result.Commands["test-cmd"]
		require.NotNil(t, cmd)
		assert.Equal(t, original.Commands["test-cmd"].Name, cmd.Name)
		assert.Equal(t, original.Commands["test-cmd"].Version, cmd.Version)
		assert.Equal(t, original.Commands["test-cmd"].Source, cmd.Source)
		assert.Equal(t, original.Commands["test-cmd"].Resolved, cmd.Resolved)
		assert.Equal(t, original.Commands["test-cmd"].Commit, cmd.Commit)
		assert.Equal(t, original.Commands["test-cmd"].InstalledAt.Unix(), cmd.InstalledAt.Unix())
		assert.Equal(t, original.Commands["test-cmd"].UpdatedAt.Unix(), cmd.UpdatedAt.Unix())
	})
}
