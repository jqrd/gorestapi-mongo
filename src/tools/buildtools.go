//go:build tools
// +build tools

// Package tools records tool dependencies. It cannot actually be compiled.
package tools

import (
	_ "github.com/srikrsna/protoc-gen-gotag"
	_ "github.com/swaggo/swag/cmd/swag"
	_ "github.com/vektra/mockery/v3"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
