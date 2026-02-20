package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListWorkspacesArgs struct {
	Pagelen int `json:"pagelen,omitempty" jsonschema:"Number of results per page (default 25, max 100)"`
	Page    int `json:"page,omitempty" jsonschema:"Page number (1-based)"`
}

// ListWorkspacesHandler returns workspaces for the authenticated user.
func (c *Client) ListWorkspacesHandler(ctx context.Context, req *mcp.CallToolRequest, args ListWorkspacesArgs) (*mcp.CallToolResult, any, error) {
	pagelen := args.Pagelen
	if pagelen == 0 {
		pagelen = 25
	}
	page := args.Page
	if page == 0 {
		page = 1
	}

	path := fmt.Sprintf("/workspaces?pagelen=%d&page=%d", pagelen, page)
	result, err := GetPaginated[Workspace](c, path)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to list workspaces: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type GetWorkspaceArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug or UUID"`
}

// GetWorkspaceHandler returns details for a single workspace.
func (c *Client) GetWorkspaceHandler(ctx context.Context, req *mcp.CallToolRequest, args GetWorkspaceArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" {
		return ToolResultError("workspace is required"), nil, nil
	}

	ws, err := GetJSON[Workspace](c, fmt.Sprintf("/workspaces/%s", QueryEscape(args.Workspace)))
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to get workspace: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(ws, "", "  ")
	return ToolResultText(string(data)), nil, nil
}
