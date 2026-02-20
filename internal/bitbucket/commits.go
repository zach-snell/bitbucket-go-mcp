package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListCommitsArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	Revision  string `json:"revision,omitempty" jsonschema:"Branch name or commit hash to list commits for"`
	Pagelen   int    `json:"pagelen,omitempty" jsonschema:"Results per page"`
	Page      int    `json:"page,omitempty" jsonschema:"Page number"`
	Include   string `json:"include,omitempty" jsonschema:"Include commits reachable from this ref"`
	Exclude   string `json:"exclude,omitempty" jsonschema:"Exclude commits reachable from this ref"`
	Path      string `json:"path,omitempty" jsonschema:"Filter commits that touch this file path"`
}

// ListCommitsHandler lists commits for a repository or branch.
func (c *Client) ListCommitsHandler(ctx context.Context, req *mcp.CallToolRequest, args ListCommitsArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" {
		return ToolResultError("workspace and repo_slug are required"), nil, nil
	}

	pagelen := args.Pagelen
	if pagelen == 0 {
		pagelen = 25
	}
	page := args.Page
	if page == 0 {
		page = 1
	}

	var endpoint string
	if args.Revision != "" {
		endpoint = fmt.Sprintf("/repositories/%s/%s/commits/%s?pagelen=%d&page=%d",
			QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), QueryEscape(args.Revision), pagelen, page)
	} else {
		endpoint = fmt.Sprintf("/repositories/%s/%s/commits?pagelen=%d&page=%d",
			QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), pagelen, page)
	}

	if args.Include != "" {
		endpoint += "&include=" + QueryEscape(args.Include)
	}
	if args.Exclude != "" {
		endpoint += "&exclude=" + QueryEscape(args.Exclude)
	}
	if args.Path != "" {
		endpoint += "&path=" + QueryEscape(args.Path)
	}

	result, err := GetPaginated[Commit](c, endpoint)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to list commits: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type GetCommitArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	Commit    string `json:"commit" jsonschema:"Commit hash"`
}

// GetCommitHandler gets a single commit by hash.
func (c *Client) GetCommitHandler(ctx context.Context, req *mcp.CallToolRequest, args GetCommitArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.Commit == "" {
		return ToolResultError("workspace, repo_slug, and commit are required"), nil, nil
	}

	c2, err := GetJSON[Commit](c, fmt.Sprintf("/repositories/%s/%s/commit/%s",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), QueryEscape(args.Commit)))
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to get commit: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(c2, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type GetDiffArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	Spec      string `json:"spec" jsonschema:"Diff spec: single commit hash or 'hash1..hash2'"`
	Path      string `json:"path,omitempty" jsonschema:"Filter diff to this file path"`
}

// GetDiffHandler gets the diff between two revisions or for a single commit.
func (c *Client) GetDiffHandler(ctx context.Context, req *mcp.CallToolRequest, args GetDiffArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.Spec == "" {
		return ToolResultError("workspace, repo_slug, and spec are required"), nil, nil
	}

	endpoint := fmt.Sprintf("/repositories/%s/%s/diff/%s",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.Spec)
	if args.Path != "" {
		endpoint += "?path=" + args.Path
	}

	raw, _, err := c.GetRaw(endpoint)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to get diff: %v", err)), nil, nil
	}

	return ToolResultText(string(raw)), nil, nil
}

type GetDiffStatArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	Spec      string `json:"spec" jsonschema:"Diff spec: single commit hash or 'hash1..hash2'"`
}

// GetDiffStatHandler gets the diff stat for a revision spec.
func (c *Client) GetDiffStatHandler(ctx context.Context, req *mcp.CallToolRequest, args GetDiffStatArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.Spec == "" {
		return ToolResultError("workspace, repo_slug, and spec are required"), nil, nil
	}

	result, err := GetPaginated[DiffStat](c, fmt.Sprintf("/repositories/%s/%s/diffstat/%s",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.Spec))
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to get diffstat: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ToolResultText(string(data)), nil, nil
}
