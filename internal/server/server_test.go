package server

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"remote-agent-mcp/internal/tools"
)

// newTestServer builds a server rooted in a temp directory and returns it.
func newTestServer(t *testing.T) *mcp.Server {
	t.Helper()
	ts := tools.NewToolSet(tools.Options{
		Root:  t.TempDir(),
		Shell: tools.ResolveShell("/bin/sh"),
	})
	return New(ts)
}

// connect wires a real MCP client to the server over in-memory transports.
// The elicitation handler accepts or declines based on the accept flag.
func connect(t *testing.T, s *mcp.Server, accept bool) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	c := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, &mcp.ClientOptions{
		ElicitationHandler: func(context.Context, *mcp.ElicitRequest) (*mcp.ElicitResult, error) {
			action := "decline"
			if accept {
				action = "accept"
			}
			return &mcp.ElicitResult{Action: action, Content: map[string]any{"confirm": accept}}, nil
		},
	})
	t1, t2 := mcp.NewInMemoryTransports()
	if _, err := s.Connect(ctx, t1, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	cs, err := c.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs
}

// resultText extracts the JSON text payload from a tool result's content.
func resultText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if len(res.Content) == 0 {
		t.Fatal("no content in result")
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("content[0] is %T, want *mcp.TextContent", res.Content[0])
	}
	return tc.Text
}

func TestE2ETools(t *testing.T) {
	s := newTestServer(t)
	cs := connect(t, s, true)
	ctx := context.Background()

	// write a new file
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "write", Arguments: map[string]any{"path": "a.txt", "content": "hello\n"}})
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if res.IsError {
		t.Fatalf("write returned error: %s", resultText(t, res))
	}

	// read it back
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: "read", Arguments: map[string]any{"path": "a.txt"}})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var readRes struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(resultText(t, res)), &readRes); err != nil {
		t.Fatalf("unmarshal read: %v", err)
	}
	if readRes.Content != "hello" {
		t.Errorf("read content = %q, want hello", readRes.Content)
	}

	// edit it (elicitation accepted)
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: "edit", Arguments: map[string]any{"path": "a.txt", "oldString": "hello", "newString": "hi"}})
	if err != nil {
		t.Fatalf("edit: %v", err)
	}
	if res.IsError {
		t.Fatalf("edit returned error: %s", resultText(t, res))
	}

	// list
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: "list", Arguments: map[string]any{}})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if res.IsError {
		t.Fatalf("list returned error: %s", resultText(t, res))
	}

	// bash
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: "bash", Arguments: map[string]any{"command": "echo ok"}})
	if err != nil {
		t.Fatalf("bash: %v", err)
	}
	if res.IsError {
		t.Fatalf("bash returned error: %s", resultText(t, res))
	}
}

func TestE2EWriteOverwriteDeclined(t *testing.T) {
	s := newTestServer(t)
	cs := connect(t, s, false) // elicitation declines
	ctx := context.Background()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "write", Arguments: map[string]any{"path": "a.txt", "content": "first\n"}})
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if res.IsError {
		t.Fatalf("initial write should succeed: %s", resultText(t, res))
	}

	// Overwriting now triggers elicitation, which declines.
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: "write", Arguments: map[string]any{"path": "a.txt", "content": "second\n"}})
	if err != nil {
		t.Fatalf("overwrite write: %v", err)
	}
	if !res.IsError {
		t.Error("overwrite with declined elicitation should be an error")
	}
}

func TestE2EPathEscapeRejected(t *testing.T) {
	s := newTestServer(t)
	cs := connect(t, s, true)
	ctx := context.Background()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "read", Arguments: map[string]any{"path": "../etc/passwd"}})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !res.IsError {
		t.Error("path escape should be rejected")
	}
}

func TestE2EDangerousCommandDeclined(t *testing.T) {
	s := newTestServer(t)
	cs := connect(t, s, false)
	ctx := context.Background()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "bash", Arguments: map[string]any{"command": "rm -rf /tmp/foo"}})
	if err != nil {
		t.Fatalf("bash: %v", err)
	}
	if !res.IsError {
		t.Error("dangerous command with declined elicitation should be an error")
	}
}

// TestE2EFilesWrittenInsideRoot verifies tools operate inside the root and the
// artifacts are actually created.
func TestE2EFilesWrittenInsideRoot(t *testing.T) {
	s := newTestServer(t)
	cs := connect(t, s, true)
	ctx := context.Background()

	_, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "write", Arguments: map[string]any{"path": filepath.Join("x", "y.txt"), "content": "z"}})
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	// The root is a temp dir; verify the file was created relative to it via list.
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "list", Arguments: map[string]any{"path": "x"}})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	text := resultText(t, res)
	if !json.Valid([]byte(text)) {
		t.Fatalf("list result is not JSON: %s", text)
	}
	var lr struct {
		Entries []struct {
			Name string `json:"name"`
		} `json:"entries"`
	}
	if err := json.Unmarshal([]byte(text), &lr); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	if len(lr.Entries) != 1 || lr.Entries[0].Name != "y.txt" {
		t.Errorf("entries = %+v, want [y.txt]", lr.Entries)
	}
}
