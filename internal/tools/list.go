package tools

import (
	"context"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listArgs struct {
	Path string `json:"path,omitempty" jsonschema:"Directory to list, relative to the root directory (default: root)"`
}

// ListEntry describes a single directory entry.
type ListEntry struct {
	Name  string `json:"name" jsonschema:"Entry name"`
	Type  string `json:"type" jsonschema:"Either \"file\" or \"dir\""`
	Size  int64  `json:"size" jsonschema:"Size in bytes"`
	MTime string `json:"mtime" jsonschema:"Modification time in RFC3339"`
}

// ListResult is the structured output of the list tool.
type ListResult struct {
	Entries []ListEntry `json:"entries" jsonschema:"Directory entries"`
}

// ListDir lists the immediate (non-recursive) contents of path, including
// hidden files.
func ListDir(path string) ([]ListEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	out := make([]ListEntry, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		typ := "file"
		if e.IsDir() {
			typ = "dir"
		}
		out = append(out, ListEntry{
			Name:  e.Name(),
			Type:  typ,
			Size:  info.Size(),
			MTime: info.ModTime().Format(time.RFC3339),
		})
	}
	return out, nil
}

func (t *ToolSet) registerList(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list",
		Description: "List the contents of a directory (single level, including hidden files).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, t.list)
}

func (t *ToolSet) list(ctx context.Context, req *mcp.CallToolRequest, args listArgs) (*mcp.CallToolResult, ListResult, error) {
	if args.Path == "" {
		args.Path = "."
	}
	full, err := ResolvePath(t.opts.Root, args.Path, t.opts.AllowAll)
	if err != nil {
		return nil, ListResult{}, err
	}
	entries, err := ListDir(full)
	if err != nil {
		return nil, ListResult{}, err
	}
	return nil, ListResult{Entries: entries}, nil
}
