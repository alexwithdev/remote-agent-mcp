package tools

import "github.com/modelcontextprotocol/go-sdk/mcp"

// Options configures a ToolSet.
type Options struct {
	Root             string
	Shell            string
	AllowAll         bool
	AllowUnconfirmed bool
}

// ToolSet holds the tool implementations and their shared configuration.
type ToolSet struct {
	opts Options
}

// NewToolSet creates a ToolSet.
func NewToolSet(opts Options) *ToolSet {
	return &ToolSet{opts: opts}
}

// Register registers all tools on the server.
func (t *ToolSet) Register(s *mcp.Server) {
	t.registerRead(s)
	t.registerWrite(s)
	t.registerEdit(s)
	t.registerList(s)
	t.registerBash(s)
}
