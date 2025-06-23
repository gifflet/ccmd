/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package main

import (
	"fmt"
	"strings"
)

// simulateVersionParsing demonstrates how version specification works
// when both @ notation and --version flag are present
func simulateVersionParsing(args []string) {
	fmt.Printf("Command: %s\n", strings.Join(args, " "))
	fmt.Println("---")

	var (
		commandName      string
		atVersion        string
		versionFlag      string
		effectiveVersion string
	)

	// Parse command arguments
	for i, arg := range args {
		// Skip program name
		if i == 0 {
			continue
		}

		// Check for @version notation
		if strings.Contains(arg, "@") && commandName == "" {
			parts := strings.SplitN(arg, "@", 2)
			commandName = parts[0]
			if len(parts) > 1 {
				atVersion = parts[1]
			}
			continue
		}

		// Check for --version flag
		if arg == "--version" && i+1 < len(args) {
			versionFlag = args[i+1]
			break
		}

		// If no @ found, first arg is command name
		if commandName == "" {
			commandName = arg
		}
	}

	// Determine which version takes precedence
	if versionFlag != "" {
		effectiveVersion = versionFlag
		fmt.Printf("Command: %s\n", commandName)
		fmt.Printf("@ version: %s\n", atVersion)
		fmt.Printf("--version flag: %s\n", versionFlag)
		fmt.Printf("✓ Using version from --version flag: %s\n", effectiveVersion)
	} else if atVersion != "" {
		effectiveVersion = atVersion
		fmt.Printf("Command: %s\n", commandName)
		fmt.Printf("@ version: %s\n", atVersion)
		fmt.Printf("--version flag: (not specified)\n")
		fmt.Printf("✓ Using version from @ notation: %s\n", effectiveVersion)
	} else {
		fmt.Printf("Command: %s\n", commandName)
		fmt.Printf("@ version: (not specified)\n")
		fmt.Printf("--version flag: (not specified)\n")
		fmt.Printf("✓ Using latest version\n")
	}

	fmt.Println()
}

func main() {
	examples := [][]string{
		// Example 1: Only @ notation
		{"ccmd", "tool@1.0.0"},

		// Example 2: Only --version flag
		{"ccmd", "tool", "--version", "2.0.0"},

		// Example 3: Both @ and --version (--version takes precedence)
		{"ccmd", "tool@1.0.0", "--version", "2.0.0"},

		// Example 4: Neither specified
		{"ccmd", "tool"},
	}

	fmt.Println("Version Specification Precedence Examples")
	fmt.Println("========================================")
	fmt.Println()

	for i, example := range examples {
		fmt.Printf("Example %d:\n", i+1)
		simulateVersionParsing(example)
	}

	// Real-world scenario
	fmt.Println("Real-world scenario:")
	fmt.Println("===================")
	fmt.Println()

	fmt.Println("Scenario: You have a script that defaults to tool@1.0.0,")
	fmt.Println("but you want to override it for testing:")
	fmt.Println()

	scriptDefault := []string{"ccmd", "mytool@1.0.0", "run", "--flag", "value"}
	fmt.Println("Script default command:")
	simulateVersionParsing(scriptDefault)

	scriptOverride := []string{"ccmd", "mytool@1.0.0", "run", "--flag", "value", "--version", "2.0.0-beta"}
	fmt.Println("Override with --version flag:")
	simulateVersionParsing(scriptOverride)

	fmt.Println("Key takeaway: --version flag always takes precedence over @ notation,")
	fmt.Println("allowing runtime overrides of hardcoded versions.")
}
