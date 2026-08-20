package tools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type editArgs struct {
	Path       string `json:"path" jsonschema:"File path to edit, relative to the root directory (or absolute if within root)"`
	OldString  string `json:"oldString" jsonschema:"Exact text to find"`
	NewString  string `json:"newString" jsonschema:"Replacement text"`
	ReplaceAll bool   `json:"replaceAll,omitempty" jsonschema:"Replace all occurrences (default false; multiple matches error unless true)"`
}

// EditResult is the structured output of the edit tool.
type EditResult struct {
	Path        string `json:"path" jsonschema:"Resolved file path"`
	Occurrences int    `json:"occurrences" jsonschema:"Number of occurrences replaced"`
	ReplaceAll  bool   `json:"replaceAll" jsonschema:"Whether replace-all mode was used"`
}

// EditFile performs a precise oldString→newString replacement in the file at
// path. It errors when oldString is empty, not found, or (without replaceAll)
// matches more than once.
func EditFile(path, oldString, newString string, replaceAll bool) (EditResult, error) {
	if oldString == "" {
		return EditResult{}, fmt.Errorf("oldString must not be empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return EditResult{}, err
	}
	content := string(data)
	count := strings.Count(content, oldString)
	if count == 0 {
		return EditResult{}, fmt.Errorf("oldString not found in file %q", path)
	}
	if !replaceAll && count > 1 {
		return EditResult{}, fmt.Errorf("oldString matches %d times; use replaceAll=true to replace all occurrences, or make oldString unique", count)
	}
	if replaceAll {
		content = strings.ReplaceAll(content, oldString, newString)
	} else {
		content = strings.Replace(content, oldString, newString, 1)
	}

	mode := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		return EditResult{}, err
	}

	occ := 1
	if replaceAll {
		occ = count
	}
	return EditResult{Path: path, Occurrences: occ, ReplaceAll: replaceAll}, nil
}

func (t *ToolSet) registerEdit(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "edit",
		Description: "Make a precise oldString→newString replacement in a file, optionally replacing all occurrences.",
	}, t.edit)
}

func (t *ToolSet) edit(ctx context.Context, req *mcp.CallToolRequest, args editArgs) (*mcp.CallToolResult, EditResult, error) {
	full, err := ResolvePath(t.opts.Root, args.Path, t.opts.AllowAll)
	if err != nil {
		return nil, EditResult{}, err
	}
	// Confirm before modifying the file, consistent with write-on-overwrite.
	if err := Confirm(ctx, req.Session, fmt.Sprintf("Modify file %q (replace %q with %q)?", full, args.OldString, args.NewString), t.opts.AllowUnconfirmed); err != nil {
		return nil, EditResult{}, err
	}
	res, err := EditFile(full, args.OldString, args.NewString, args.ReplaceAll)
	if err != nil {
		return nil, EditResult{}, err
	}
	return nil, res, nil
}
