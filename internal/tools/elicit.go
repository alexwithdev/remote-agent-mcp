package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Elicitor abstracts the elicitation capability so confirmation logic can be
// unit-tested without a live MCP session. *mcp.ServerSession satisfies it.
type Elicitor interface {
	Elicit(ctx context.Context, params *mcp.ElicitParams) (*mcp.ElicitResult, error)
}

// Confirm asks the user to approve an action via elicitation.
//
// It returns nil if the user accepts (or if elicitation is unavailable and
// allowUnconfirmed is true), and a descriptive error otherwise.
func Confirm(ctx context.Context, e Elicitor, message string, allowUnconfirmed bool) error {
	if e == nil {
		if allowUnconfirmed {
			return nil
		}
		return errors.New("elicitation is unavailable: client does not support elicitation (set --allow-unconfirmed to skip confirmation)")
	}
	res, err := e.Elicit(ctx, &mcp.ElicitParams{
		Message: message,
		RequestedSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"confirm": {Type: "boolean", Description: "Approve this action?"},
			},
			Required: []string{"confirm"},
		},
	})
	if err != nil {
		if allowUnconfirmed {
			return nil
		}
		return fmt.Errorf("elicitation failed (client may not support it): %w (set --allow-unconfirmed to skip confirmation)", err)
	}
	switch res.Action {
	case "accept":
		if confirmed, ok := res.Content["confirm"].(bool); ok && confirmed {
			return nil
		}
		return errors.New("user did not confirm the action")
	case "decline":
		return errors.New("user declined the action")
	case "cancel":
		return errors.New("user cancelled the confirmation")
	default:
		return fmt.Errorf("unexpected elicitation action %q", res.Action)
	}
}
