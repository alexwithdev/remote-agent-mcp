// Package logging builds the server's structured logger: level-filtered,
// dual-written to a file and stderr, with size-based rotation.
package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Config holds the logging configuration.
type Config struct {
	// Level is the minimum level to emit: debug, info, warn, or error.
	Level string
	// File is the log file path. When empty, logs go only to stderr.
	File string
	// Stderr is the writer for the stderr stream. When nil, os.Stderr is used.
	Stderr io.Writer
	// MaxSize is the maximum size of a single log file in megabytes before
	// rotation. Only used when File is non-empty.
	MaxSize int
	// MaxBackups is the maximum number of rotated log files to retain. Only
	// used when File is non-empty.
	MaxBackups int
}

// New builds a *slog.Logger that writes to stderr and, when File is set, to a
// rotating file. The returned closer releases the file writer; callers must
// call it on shutdown.
func New(cfg Config) (*slog.Logger, io.Closer, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, nil, err
	}

	stderr := cfg.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	// Build the file writer (rotation) when a file path is configured.
	var fileWriter io.Writer
	var closer io.Closer = nopCloser{}
	if cfg.File != "" {
		lj := newRotator(cfg.File, cfg.MaxSize, cfg.MaxBackups)
		fileWriter = lj
		closer = lj
	}

	// Combine file and stderr into a single multi-writer.
	var out io.Writer = stderr
	if fileWriter != nil {
		out = io.MultiWriter(stderr, fileWriter)
	}

	handler := slog.NewTextHandler(out, &slog.HandlerOptions{Level: level})
	return slog.New(handler), closer, nil
}

func parseLevel(s string) (slog.Level, error) {
	switch s {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid log level %q (want debug, info, warn, or error)", s)
	}
}

type nopCloser struct{}

func (nopCloser) Close() error { return nil }

// WrapTool wraps an MCP tool handler so every invocation is logged at debug
// level with the tool name, its arguments, the call duration in milliseconds,
// and whether it succeeded (plus the error when it failed).
func WrapTool[In, Out any](logger *slog.Logger, name string, h mcp.ToolHandlerFor[In, Out]) mcp.ToolHandlerFor[In, Out] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input In) (*mcp.CallToolResult, Out, error) {
		start := time.Now()
		result, out, err := h(ctx, req, input)
		duration := time.Since(start).Milliseconds()

		args := argsGroup(input)
		if err != nil {
			logger.Debug("tool call",
				"tool", name,
				"args", args,
				"duration_ms", duration,
				"ok", false,
				"error", err.Error(),
			)
		} else {
			logger.Debug("tool call",
				"tool", name,
				"args", args,
				"duration_ms", duration,
				"ok", true,
			)
		}
		return result, out, err
	}
}

// argsGroup renders the tool arguments as a slog group with JSON-style field
// names, so text handlers show args.path=/tmp/x rather than a compact JSON
// blob. It falls back to the raw value when it cannot be JSON-encoded.
func argsGroup[In any](input In) slog.Value {
	data, err := json.Marshal(input)
	if err != nil {
		return slog.AnyValue(input)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return slog.AnyValue(input)
	}
	attrs := make([]slog.Attr, 0, len(m))
	for k, v := range m {
		attrs = append(attrs, slog.Any(k, v))
	}
	return slog.GroupValue(attrs...)
}
