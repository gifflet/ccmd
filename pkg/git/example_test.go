/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package git_test

import (
	"fmt"
	"log"

	"github.com/gifflet/ccmd/pkg/git"
)

func ExampleClient_Clone() {
	// Create a new Git client without authentication
	client, err := git.New(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Clone a repository
	err = client.Clone("https://github.com/user/repo.git", "/tmp/repo")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Repository cloned successfully")
}

func ExampleClient_Clone_withAuth() {
	// Create a Git client with basic authentication
	client, err := git.New(&git.Options{
		Username: "username",
		Password: "password",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Clone a private repository
	err = client.Clone("https://github.com/user/private-repo.git", "/tmp/private-repo")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Private repository cloned successfully")
}

func ExampleClient_ListTags() {
	client, err := git.New(nil)
	if err != nil {
		log.Fatal(err)
	}

	// List all tags in a repository
	tags, err := client.ListTags("/path/to/repo")
	if err != nil {
		log.Fatal(err)
	}

	for _, tag := range tags {
		fmt.Println(tag)
	}
}

func ExampleClient_Checkout() {
	client, err := git.New(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Checkout a specific tag
	err = client.Checkout("/path/to/repo", "v1.0.0")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Checked out tag v1.0.0")
}

func ExampleGetRepositoryName() {
	// Extract repository name from various URL formats
	urls := []string{
		"https://github.com/user/repo.git",
		"git@github.com:user/repo.git",
		"/path/to/local/repo",
		"user/repo",
	}

	for _, url := range urls {
		name := git.GetRepositoryName(url)
		fmt.Printf("%s -> %s\n", url, name)
	}
	// Output:
	// https://github.com/user/repo.git -> repo
	// git@github.com:user/repo.git -> repo
	// /path/to/local/repo -> repo
	// user/repo -> repo
}
