package tools

import (
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"remote-agent-mcp/internal/logging"
)

// Options configures a ToolSet.
type Options struct {
	Root             string
	Shell            string
	AllowAll         bool
	AllowUnconfirmed bool
	// Logger, when non-nil, is used to log tool calls at debug level.
	Logger *slog.Logger
}

// ToolSet holds the tool implementations and their shared configuration.
type ToolSet struct {
	opts   Options
	logger *slog.Logger
}

// NewToolSet creates a ToolSet.
func NewToolSet(opts Options) *ToolSet {
	return &ToolSet{opts: opts, logger: opts.Logger}
}

// Register registers all tools on the server.
func (t *ToolSet) Register(s *mcp.Server) {
	t.registerRead(s)
	t.registerWrite(s)
	t.registerEdit(s)
	t.registerList(s)
	t.registerBash(s)
}

// wrapTool wraps a tool handler so its invocations are logged at debug level.
// When logger is nil, the handler is returned unchanged.
func wrapTool[In, Out any](logger *slog.Logger, name string, h mcp.ToolHandlerFor[In, Out]) mcp.ToolHandlerFor[In, Out] {
	if logger == nil {
		return h
	}
	return logging.WrapTool(logger, name, h)
}
