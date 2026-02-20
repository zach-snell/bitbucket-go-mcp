package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// ListCommitsHandler lists commits for a repository or branch.
func (c *Client) ListCommitsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	revision := req.GetString("revision", "")
	pagelen := req.GetInt("pagelen", 25)
	page := req.GetInt("page", 1)
	include := req.GetString("include", "")
	exclude := req.GetString("exclude", "")
	path := req.GetString("path", "")

	var endpoint string
	if revision != "" {
		endpoint = fmt.Sprintf("/repositories/%s/%s/commits/%s?pagelen=%d&page=%d",
			QueryEscape(workspace), QueryEscape(repoSlug), QueryEscape(revision), pagelen, page)
	} else {
		endpoint = fmt.Sprintf("/repositories/%s/%s/commits?pagelen=%d&page=%d",
			QueryEscape(workspace), QueryEscape(repoSlug), pagelen, page)
	}

	if include != "" {
		endpoint += "&include=" + include
	}
	if exclude != "" {
		endpoint += "&exclude=" + exclude
	}
	if path != "" {
		endpoint += "&path=" + path
	}

	result, err := GetPaginated[Commit](c, endpoint)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list commits: %v", err)), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// GetCommitHandler gets a single commit by hash.
func (c *Client) GetCommitHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	commit, err := req.RequireString("commit")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	c2, err := GetJSON[Commit](c, fmt.Sprintf("/repositories/%s/%s/commit/%s",
		QueryEscape(workspace), QueryEscape(repoSlug), QueryEscape(commit)))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get commit: %v", err)), nil
	}

	data, _ := json.MarshalIndent(c2, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// GetDiffHandler gets the diff between two revisions or for a single commit.
func (c *Client) GetDiffHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	spec, err := req.RequireString("spec")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	diffPath := req.GetString("path", "")

	endpoint := fmt.Sprintf("/repositories/%s/%s/diff/%s",
		QueryEscape(workspace), QueryEscape(repoSlug), spec)
	if diffPath != "" {
		endpoint += "?path=" + diffPath
	}

	raw, _, err := c.GetRaw(endpoint)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get diff: %v", err)), nil
	}

	return mcp.NewToolResultText(string(raw)), nil
}

// GetDiffStatHandler gets the diff stat for a revision spec.
func (c *Client) GetDiffStatHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	spec, err := req.RequireString("spec")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result, err := GetPaginated[DiffStat](c, fmt.Sprintf("/repositories/%s/%s/diffstat/%s",
		QueryEscape(workspace), QueryEscape(repoSlug), spec))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get diffstat: %v", err)), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
