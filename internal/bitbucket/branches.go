package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// ListBranchesHandler lists branches in a repository.
func (c *Client) ListBranchesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	pagelen := req.GetInt("pagelen", 25)
	page := req.GetInt("page", 1)
	query := req.GetString("query", "")
	sort := req.GetString("sort", "")

	path := fmt.Sprintf("/repositories/%s/%s/refs/branches?pagelen=%d&page=%d",
		QueryEscape(workspace), QueryEscape(repoSlug), pagelen, page)
	if query != "" {
		path += "&q=" + query
	}
	if sort != "" {
		path += "&sort=" + sort
	}

	result, err := GetPaginated[Branch](c, path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list branches: %v", err)), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// CreateBranchHandler creates a new branch from a commit hash.
func (c *Client) CreateBranchHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	target, err := req.RequireString("target")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	body := CreateBranchRequest{
		Name:   name,
		Target: map[string]string{"hash": target},
	}

	respData, err := c.Post(fmt.Sprintf("/repositories/%s/%s/refs/branches",
		QueryEscape(workspace), QueryEscape(repoSlug)), body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create branch: %v", err)), nil
	}

	var branch Branch
	if err := json.Unmarshal(respData, &branch); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse response: %v", err)), nil
	}

	data, _ := json.MarshalIndent(branch, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// DeleteBranchHandler deletes a branch.
func (c *Client) DeleteBranchHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := c.Delete(fmt.Sprintf("/repositories/%s/%s/refs/branches/%s",
		QueryEscape(workspace), QueryEscape(repoSlug), QueryEscape(name))); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to delete branch: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Branch '%s' deleted successfully", name)), nil
}

// ListTagsHandler lists tags in a repository.
func (c *Client) ListTagsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	pagelen := req.GetInt("pagelen", 25)
	page := req.GetInt("page", 1)

	path := fmt.Sprintf("/repositories/%s/%s/refs/tags?pagelen=%d&page=%d",
		QueryEscape(workspace), QueryEscape(repoSlug), pagelen, page)

	result, err := GetPaginated[Tag](c, path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list tags: %v", err)), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// CreateTagHandler creates a new tag.
func (c *Client) CreateTagHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	target, err := req.RequireString("target")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	body := map[string]interface{}{
		"name":   name,
		"target": map[string]string{"hash": target},
	}

	respData, err := c.Post(fmt.Sprintf("/repositories/%s/%s/refs/tags",
		QueryEscape(workspace), QueryEscape(repoSlug)), body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create tag: %v", err)), nil
	}

	var tag Tag
	if err := json.Unmarshal(respData, &tag); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse response: %v", err)), nil
	}

	data, _ := json.MarshalIndent(tag, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
