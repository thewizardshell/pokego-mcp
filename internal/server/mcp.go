// Package server provides MCP (Model Context Protocol) server implementation.
// It wraps the mcp-go server with additional functionality for Pokemon tools.
package server

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/server"
)

// MCPServer wraps the underlying MCP server with metadata and convenience methods.
type MCPServer struct {
	Server  *server.MCPServer
	Name    string
	Version string
}

// NewMCPServer creates a new MCP server instance with the specified name and version.
// It configures the server with tool capabilities, recovery middleware, and logging.
func NewMCPServer(name string, version string) *MCPServer {
	srv := server.NewMCPServer(
		name,
		version,
		server.WithToolCapabilities(true),
		server.WithRecovery(),
		server.WithLogging(),
	)

	return &MCPServer{
		Server:  srv,
		Name:    name,
		Version: version,
	}
}

// StartSSE starts the MCP server in Server-Sent Events (SSE) mode.
// This mode is useful for web-based clients that can consume SSE streams.
// The server will listen on the specified port with /sse and /message endpoints.
func (m *MCPServer) StartSSE(ctx context.Context, port int) error {
	if m.Server == nil {
		return fmt.Errorf("MCP server is not initialized")
	}

	baseURL := fmt.Sprintf("http://localhost:%d", port)

	sseServer := server.NewSSEServer(
		m.Server,
		server.WithBaseURL(baseURL),
		server.WithSSEEndpoint("/sse"),
		server.WithMessageEndpoint("/message"),
	)

	addr := fmt.Sprintf(":%d", port)
	if err := sseServer.Start(addr); err != nil {
		return fmt.Errorf("failed to start SSE server: %w", err)
	}

	return nil
}

// StartSdio starts the MCP server in STDIO mode.
// This mode uses standard input/output for communication and is primarily
// designed for integration with Claude Desktop and similar applications.
func (m *MCPServer) StartSdio(ctx context.Context) error {
	if m.Server == nil {
		return fmt.Errorf("MCP server is not initialized")
	}

	if err := server.ServeStdio(m.Server); err != nil {
		return fmt.Errorf("failed to start STDIO server: %w", err)
	}
	return nil
}
