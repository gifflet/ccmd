// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package errors_test

import (
	"fmt"
	"os"

	"github.com/gifflet/ccmd/pkg/errors"
	"github.com/gifflet/ccmd/pkg/logger"
)

func ExampleError_basic() {
	// Create a simple error
	err := errors.New(errors.CodeNotFound, "command not found")
	fmt.Println(err)
	// Output: [NOT_FOUND] command not found
}

func ExampleError_withDetails() {
	// Create an error with details
	err := errors.New(errors.CodeConfigInvalid, "invalid configuration").
		WithDetail("file", "ccmd.yaml").
		WithDetail("line", 42)

	// The error includes the code and message
	fmt.Println(err)
	// Output: [CONFIG_INVALID] invalid configuration
}

func ExampleWrap() {
	// Simulate an underlying error
	originalErr := fmt.Errorf("connection timeout")

	// Wrap it with context
	err := errors.Wrap(originalErr, errors.CodeNetworkTimeout, "failed to fetch repository")

	fmt.Println(err)
	// Output: [NETWORK_TIMEOUT] failed to fetch repository: connection timeout
}

func ExampleHandler() {
	// Set up a test logger
	log := logger.New(os.Stdout, logger.InfoLevel)
	handler := errors.NewHandler(log)

	// Create an error
	err := errors.New(errors.CodeCommandNotFound, "command 'test' not found").
		WithDetail("command", "test")

	// Handle the error (this would normally print to stderr)
	handler.Handle(err)
}

func ExampleIsNotFound() {
	err1 := errors.New(errors.CodeNotFound, "not found")
	err2 := errors.New(errors.CodeCommandNotFound, "command not found")
	err3 := errors.New(errors.CodeInternal, "internal error")

	fmt.Println(errors.IsNotFound(err1)) // true
	fmt.Println(errors.IsNotFound(err2)) // true
	fmt.Println(errors.IsNotFound(err3)) // false
	// Output:
	// true
	// true
	// false
}

func ExampleGetCode() {
	// Create different types of errors
	err1 := errors.New(errors.CodeNotFound, "not found")
	err2 := fmt.Errorf("standard error")

	// Get error codes
	fmt.Println(errors.GetCode(err1))
	fmt.Println(errors.GetCode(err2))
	// Output:
	// NOT_FOUND
	// UNKNOWN
}
