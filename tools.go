//go:build tools
// +build tools

package main

// This is require in go version prior to v 1.24
// currently the toolset base image has go 1.23
// after migrating to go 1.24 we can remove this file and include the tool in the go.mod file directly

import (
	// for the go:generate command
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)
