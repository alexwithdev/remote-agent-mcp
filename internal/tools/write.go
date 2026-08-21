package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type writeArgs struct {
	Path    string `json:"path" jsonschema:"File path to write, relative to the root directory (or absolute if within root)"`
	Content string `json:"content" jsonschema:"Full content to write to the file"`
}

// WriteResult is the structured output of the write tool.
type WriteResult struct {
	Path         string `json:"path" jsonschema:"Resolved file path"`
	BytesWritten int    `json:"bytesWritten" jsonschema:"Number of bytes written"`
	Created      bool   `json:"created" jsonschema:"True when a new file was created, false when overwritten"`
}

// WriteFile writes content to path, creating parent directories and preserving
// the existing file mode when overwriting.
func WriteFile(path, content string) (WriteResult, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return WriteResult{}, err
	}
	mode := os.FileMode(0o644)
	created := true
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
		created = false
	}
	data := []byte(content)
	if err := os.WriteFile(path, data, mode); err != nil {
		return WriteResult{}, err
	}
	return WriteResult{Path: path, BytesWritten: len(data), Created: created}, nil
}

func (t *ToolSet) registerWrite(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "write",
		Description: "Create or overwrite a file, creating parent directories as needed.",
	}, wrapTool(t.logger, "write", t.write))
}

func (t *ToolSet) write(ctx context.Context, req *mcp.CallToolRequest, args writeArgs) (*mcp.CallToolResult, WriteResult, error) {
	full, err := ResolvePath(t.opts.Root, args.Path, t.opts.AllowAll)
	if err != nil {
		return nil, WriteResult{}, err
	}
	// Confirm before overwriting an existing file.
	if _, err := os.Stat(full); err == nil {
		if err := Confirm(ctx, req.Session, fmt.Sprintf("Overwrite existing file %q?", full), t.opts.AllowUnconfirmed); err != nil {
			return nil, WriteResult{}, err
		}
	}
	res, err := WriteFile(full, args.Content)
	if err != nil {
		return nil, WriteResult{}, err
	}
	return nil, res, nil
}
