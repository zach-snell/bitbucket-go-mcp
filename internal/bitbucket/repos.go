package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// ListRepositoriesHandler lists repositories in a workspace.
func (c *Client) ListRepositoriesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	pagelen := req.GetInt("pagelen", 25)
	page := req.GetInt("page", 1)
	query := req.GetString("query", "")
	role := req.GetString("role", "")
	sort := req.GetString("sort", "")

	path := fmt.Sprintf("/repositories/%s?pagelen=%d&page=%d", QueryEscape(workspace), pagelen, page)
	if query != "" {
		path += "&q=" + query
	}
	if role != "" {
		path += "&role=" + role
	}
	if sort != "" {
		path += "&sort=" + sort
	}

	result, err := GetPaginated[Repository](c, path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list repositories: %v", err)), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// GetRepositoryHandler gets details for a single repository.
func (c *Client) GetRepositoryHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	repo, err := GetJSON[Repository](c, fmt.Sprintf("/repositories/%s/%s",
		QueryEscape(workspace), QueryEscape(repoSlug)))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get repository: %v", err)), nil
	}

	data, _ := json.MarshalIndent(repo, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// CreateRepositoryHandler creates a new repository in a workspace.
func (c *Client) CreateRepositoryHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	body := map[string]interface{}{
		"scm": "git",
	}

	if desc := req.GetString("description", ""); desc != "" {
		body["description"] = desc
	}
	if lang := req.GetString("language", ""); lang != "" {
		body["language"] = lang
	}
	isPrivate := req.GetBool("is_private", true)
	body["is_private"] = isPrivate

	if project := req.GetString("project_key", ""); project != "" {
		body["project"] = map[string]string{"key": project}
	}

	respData, err := c.Post(fmt.Sprintf("/repositories/%s/%s",
		QueryEscape(workspace), QueryEscape(repoSlug)), body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create repository: %v", err)), nil
	}

	var repo Repository
	if err := json.Unmarshal(respData, &repo); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse response: %v", err)), nil
	}

	data, _ := json.MarshalIndent(repo, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// DeleteRepositoryHandler deletes a repository.
func (c *Client) DeleteRepositoryHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := c.Delete(fmt.Sprintf("/repositories/%s/%s",
		QueryEscape(workspace), QueryEscape(repoSlug))); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to delete repository: %v", err)), nil
	}

	return mcp.NewToolResultText("Repository deleted successfully"), nil
}
