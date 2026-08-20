// Package server builds and serves the MCP server over Streamable HTTP.
package server

import (
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"remote-agent-mcp/internal/tools"
)

const (
	name    = "remote-agent-mcp"
	version = "0.1.0"
)

// New builds the MCP server with all tools registered.
func New(ts *tools.ToolSet) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    name,
		Version: version,
	}, &mcp.ServerOptions{
		Instructions: "Remote machine troubleshooting: read, write, edit, and list files, and run shell commands, within the configured root directory.",
	})
	ts.Register(s)
	return s
}

// Handler returns an HTTP handler serving the MCP server over Streamable HTTP,
// optionally enforcing bearer-token authentication.
func Handler(s *mcp.Server, token string) http.Handler {
	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return s
	}, nil)
	if token == "" {
		return mcpHandler
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !authorized(r, token) {
			w.Header().Set("WWW-Authenticate", `Bearer`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		mcpHandler.ServeHTTP(w, r)
	})
}

func authorized(r *http.Request, token string) bool {
	auth := r.Header.Get("Authorization")
	return strings.HasPrefix(auth, "Bearer ") && strings.TrimPrefix(auth, "Bearer ") == token
}
