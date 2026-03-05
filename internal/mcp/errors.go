// Package mcp provides MCP (Model Context Protocol) client integration
// for connecting to external MCP servers and adapting their tools.
package mcp

import "errors"

var (
	// ErrServerNotFound indicates the named MCP server is not configured.
	ErrServerNotFound = errors.New("mcp: server not found")

	// ErrConnectionFailed indicates a connection attempt failed.
	ErrConnectionFailed = errors.New("mcp: connection failed")

	// ErrToolCallFailed indicates a tool call returned an error.
	ErrToolCallFailed = errors.New("mcp: tool call failed")

	// ErrNotConnected indicates the server is not connected.
	ErrNotConnected = errors.New("mcp: not connected")

	// ErrInvalidTransport indicates an unsupported transport type.
	ErrInvalidTransport = errors.New("mcp: invalid transport type")
)
