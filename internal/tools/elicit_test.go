package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// fakeElicitor lets the confirmation logic be tested without a live session.
type fakeElicitor struct {
	action  string
	confirm bool
	err     error
}

func (f *fakeElicitor) Elicit(ctx context.Context, params *mcp.ElicitParams) (*mcp.ElicitResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &mcp.ElicitResult{Action: f.action, Content: map[string]any{"confirm": f.confirm}}, nil
}

func TestConfirmAccept(t *testing.T) {
	if err := Confirm(context.Background(), &fakeElicitor{action: "accept", confirm: true}, "ok?", false); err != nil {
		t.Errorf("accept should confirm: %v", err)
	}
}

func TestConfirmAcceptButUnconfirmed(t *testing.T) {
	if err := Confirm(context.Background(), &fakeElicitor{action: "accept", confirm: false}, "ok?", false); err == nil {
		t.Error("accept with confirm=false should return an error")
	}
}

func TestConfirmDecline(t *testing.T) {
	if err := Confirm(context.Background(), &fakeElicitor{action: "decline"}, "ok?", false); err == nil {
		t.Error("decline should return an error")
	}
}

func TestConfirmCancel(t *testing.T) {
	if err := Confirm(context.Background(), &fakeElicitor{action: "cancel"}, "ok?", false); err == nil {
		t.Error("cancel should return an error")
	}
}

func TestConfirmUnsupportedDenied(t *testing.T) {
	e := &fakeElicitor{err: errors.New("unsupported")}
	if err := Confirm(context.Background(), e, "ok?", false); err == nil {
		t.Error("unsupported elicitation without allowUnconfirmed should fail")
	}
}

func TestConfirmUnsupportedAllowed(t *testing.T) {
	e := &fakeElicitor{err: errors.New("unsupported")}
	if err := Confirm(context.Background(), e, "ok?", true); err != nil {
		t.Errorf("unsupported elicitation with allowUnconfirmed should pass: %v", err)
	}
}
