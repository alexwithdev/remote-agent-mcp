package tools

import (
	"context"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultReadLimit = 2000
	maxReadLimit     = 2000
	maxReadBytes     = 50 * 1024
)

type readArgs struct {
	Path   string `json:"path" jsonschema:"File path, relative to the root directory (or absolute if within root)"`
	Offset int    `json:"offset,omitempty" jsonschema:"1-based starting line number (default 1)"`
	Limit  int    `json:"limit,omitempty" jsonschema:"Maximum number of lines to return (default 2000, max 2000)"`
}

// ReadResult is the structured output of the read tool.
type ReadResult struct {
	Path       string `json:"path" jsonschema:"Resolved file path"`
	Content    string `json:"content" jsonschema:"Requested portion of the file"`
	Offset     int    `json:"offset" jsonschema:"Starting line number used"`
	Limit      int    `json:"limit" jsonschema:"Line limit used"`
	TotalLines int    `json:"totalLines" jsonschema:"Total number of lines in the file"`
	TotalBytes int64  `json:"totalBytes" jsonschema:"Total size of the file in bytes"`
	Truncated  bool   `json:"truncated" jsonschema:"True when the returned content is truncated"`
}

// ReadFile reads a window of the file at path using a 1-based line offset/limit.
func ReadFile(path string, offset, limit int) (ReadResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ReadResult{}, err
	}
	totalBytes := int64(len(data))
	lines := splitLines(string(data))
	totalLines := len(lines)

	if offset < 1 {
		offset = 1
	}
	if limit < 1 {
		limit = defaultReadLimit
	}
	if limit > maxReadLimit {
		limit = maxReadLimit
	}

	res := ReadResult{
		Path:       path,
		Offset:     offset,
		Limit:      limit,
		TotalLines: totalLines,
		TotalBytes: totalBytes,
	}

	start := offset - 1
	if start >= totalLines {
		return res, nil // past EOF
	}
	end := start + limit
	truncated := end < totalLines
	if end > totalLines {
		end = totalLines
	}
	content := strings.Join(lines[start:end], "\n")

	if c, t := Truncate(content, maxReadBytes); t {
		content = c
		truncated = true
	}
	res.Content = content
	res.Truncated = truncated
	return res, nil
}

// splitLines splits data on newlines, dropping a trailing empty element so
// "a\nb\n" counts as two lines.
func splitLines(data string) []string {
	s := strings.Split(data, "\n")
	if len(s) > 0 && s[len(s)-1] == "" {
		s = s[:len(s)-1]
	}
	return s
}

func (t *ToolSet) registerRead(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "read",
		Description: "Read a file's content with line-based offset/limit paging and automatic truncation.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, t.read)
}

func (t *ToolSet) read(ctx context.Context, req *mcp.CallToolRequest, args readArgs) (*mcp.CallToolResult, ReadResult, error) {
	full, err := ResolvePath(t.opts.Root, args.Path, t.opts.AllowAll)
	if err != nil {
		return nil, ReadResult{}, err
	}
	res, err := ReadFile(full, args.Offset, args.Limit)
	if err != nil {
		return nil, ReadResult{}, err
	}
	return nil, res, nil
}
