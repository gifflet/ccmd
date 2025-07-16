/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package output

import (
	"errors"
	"testing"

	ccmderrors "github.com/gifflet/ccmd/pkg/errors"
)

func TestPrintError(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "not found error",
			err:  ccmderrors.NotFound("command test"),
		},
		{
			name: "already exists error",
			err:  ccmderrors.AlreadyExists("command test"),
		},
		{
			name: "invalid input error",
			err:  ccmderrors.InvalidInput("invalid format"),
		},
		{
			name: "file operation error",
			err:  ccmderrors.FileError("read", "/path/to/file", nil),
		},
		{
			name: "git operation error",
			err:  ccmderrors.GitError("clone", nil),
		},
		{
			name: "generic error",
			err:  errors.New("generic error"),
		},
		{
			name: "nil error",
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just ensure no panic occurs
			PrintError(tt.err)
		})
	}
}
