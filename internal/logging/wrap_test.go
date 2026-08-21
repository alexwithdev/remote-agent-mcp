package logging

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type wrapArgs struct {
	Path string `json:"path"`
}

type wrapResult struct {
	OK bool `json:"ok"`
}

func TestWrapToolLogsSuccess(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := WrapTool(logger, "read", func(ctx context.Context, req *mcp.CallToolRequest, args wrapArgs) (*mcp.CallToolResult, wrapResult, error) {
		return nil, wrapResult{OK: true}, nil
	})

	_, _, err := handler(context.Background(), &mcp.CallToolRequest{}, wrapArgs{Path: "/tmp/x"})
	if err != nil {
		t.Fatalf("handler: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"tool=read", "path", "/tmp/x", "ok=true"} {
		if !strings.Contains(out, want) {
			t.Errorf("output %q missing %q", out, want)
		}
	}
	if !strings.Contains(out, "duration_ms=") {
		t.Errorf("output %q missing duration_ms", out)
	}
}

func TestWrapToolLogsError(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := WrapTool(logger, "bash", func(ctx context.Context, req *mcp.CallToolRequest, args wrapArgs) (*mcp.CallToolResult, wrapResult, error) {
		return nil, wrapResult{}, errors.New("boom")
	})

	_, _, err := handler(context.Background(), &mcp.CallToolRequest{}, wrapArgs{Path: "/tmp/x"})
	if err == nil {
		t.Fatal("expected handler error to propagate")
	}

	out := buf.String()
	for _, want := range []string{"tool=bash", "ok=false", "error=boom"} {
		if !strings.Contains(out, want) {
			t.Errorf("output %q missing %q", out, want)
		}
	}
}

func TestWrapToolPropagatesResult(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := WrapTool(logger, "read", func(ctx context.Context, req *mcp.CallToolRequest, args wrapArgs) (*mcp.CallToolResult, wrapResult, error) {
		return nil, wrapResult{OK: true}, nil
	})

	_, out, err := handler(context.Background(), &mcp.CallToolRequest{}, wrapArgs{Path: "/tmp/x"})
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !out.OK {
		t.Error("result should propagate through wrapper")
	}
}
