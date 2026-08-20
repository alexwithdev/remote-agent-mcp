// Package tools implements the MCP tools (read, write, edit, list, bash) for
// remote machine troubleshooting.
package tools

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ResolvePath resolves a user-supplied path against the root directory,
// enforcing containment unless allowAll is true.
//
// Relative paths are joined to root. Absolute paths are allowed only when they
// resolve inside root (or when allowAll is true).
//
// Containment is lexical (Abs/Rel), so it does not follow symbolic links: a
// symlink inside root pointing outside root is not traversed here. Callers
// that read/write through such links do so at their own risk.
func ResolvePath(root, p string, allowAll bool) (string, error) {
	if p == "" {
		return "", fmt.Errorf("path must not be empty")
	}
	if allowAll {
		if filepath.IsAbs(p) {
			return filepath.Clean(p), nil
		}
		return filepath.Join(root, p), nil
	}

	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolving root: %w", err)
	}
	full := p
	if !filepath.IsAbs(full) {
		full = filepath.Join(rootAbs, full)
	}
	fullAbs, err := filepath.Abs(full)
	if err != nil {
		return "", fmt.Errorf("resolving path: %w", err)
	}
	rel, err := filepath.Rel(rootAbs, fullAbs)
	if err != nil {
		return "", fmt.Errorf("relativizing path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes root %q", p, rootAbs)
	}
	return fullAbs, nil
}
