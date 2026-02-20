package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// ListWorkspacesHandler returns workspaces for the authenticated user.
func (c *Client) ListWorkspacesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pagelen := req.GetInt("pagelen", 25)
	page := req.GetInt("page", 1)

	path := fmt.Sprintf("/workspaces?pagelen=%d&page=%d", pagelen, page)
	result, err := GetPaginated[Workspace](c, path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list workspaces: %v", err)), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// GetWorkspaceHandler returns details for a single workspace.
func (c *Client) GetWorkspaceHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	ws, err := GetJSON[Workspace](c, fmt.Sprintf("/workspaces/%s", QueryEscape(workspace)))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get workspace: %v", err)), nil
	}

	data, _ := json.MarshalIndent(ws, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
