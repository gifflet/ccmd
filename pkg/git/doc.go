/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

// Package git provides a simple wrapper around go-git for common Git operations.
// It supports cloning repositories, fetching updates, checking out specific versions,
// and listing tags. The package handles various Git URL formats including GitHub
// shorthand notation and supports authentication via HTTP basic auth or SSH keys.
//
// Basic usage:
//
//	client, err := git.New(nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	err = client.Clone("user/repo", "/tmp/repo")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// With authentication:
//
//	client, err := git.New(&git.Options{
//	    Username: "username",
//	    Password: "token",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
package git
