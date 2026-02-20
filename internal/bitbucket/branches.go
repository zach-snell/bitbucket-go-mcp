package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListBranchesArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	Pagelen   int    `json:"pagelen,omitempty" jsonschema:"Results per page"`
	Page      int    `json:"page,omitempty" jsonschema:"Page number"`
	Query     string `json:"query,omitempty" jsonschema:"Filter query"`
	Sort      string `json:"sort,omitempty" jsonschema:"Sort field"`
}

// ListBranchesHandler lists branches in a repository.
func (c *Client) ListBranchesHandler(ctx context.Context, req *mcp.CallToolRequest, args ListBranchesArgs) (*mcp.CallToolResult, any, error) {
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

	path := fmt.Sprintf("/repositories/%s/%s/refs/branches?pagelen=%d&page=%d",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), pagelen, page)
	if args.Query != "" {
		path += "&q=" + QueryEscape(args.Query)
	}
	if args.Sort != "" {
		path += "&sort=" + QueryEscape(args.Sort)
	}

	result, err := GetPaginated[Branch](c, path)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to list branches: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type CreateBranchArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	Name      string `json:"name" jsonschema:"Branch name"`
	Target    string `json:"target" jsonschema:"Target commit hash to branch from"`
}

// CreateBranchHandler creates a new branch from a commit hash.
func (c *Client) CreateBranchHandler(ctx context.Context, req *mcp.CallToolRequest, args CreateBranchArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.Name == "" || args.Target == "" {
		return ToolResultError("workspace, repo_slug, name, and target are required"), nil, nil
	}

	body := CreateBranchRequest{
		Name:   args.Name,
		Target: map[string]string{"hash": args.Target},
	}

	respData, err := c.Post(fmt.Sprintf("/repositories/%s/%s/refs/branches",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug)), body)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to create branch: %v", err)), nil, nil
	}

	var branch Branch
	if err := json.Unmarshal(respData, &branch); err != nil {
		return ToolResultError(fmt.Sprintf("failed to parse response: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(branch, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type DeleteBranchArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	Name      string `json:"name" jsonschema:"Branch name to delete"`
}

// DeleteBranchHandler deletes a branch.
func (c *Client) DeleteBranchHandler(ctx context.Context, req *mcp.CallToolRequest, args DeleteBranchArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.Name == "" {
		return ToolResultError("workspace, repo_slug, and name are required"), nil, nil
	}

	if err := c.Delete(fmt.Sprintf("/repositories/%s/%s/refs/branches/%s",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), QueryEscape(args.Name))); err != nil {
		return ToolResultError(fmt.Sprintf("failed to delete branch: %v", err)), nil, nil
	}

	return ToolResultText(fmt.Sprintf("Branch '%s' deleted successfully", args.Name)), nil, nil
}

type ListTagsArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	Pagelen   int    `json:"pagelen,omitempty" jsonschema:"Results per page"`
	Page      int    `json:"page,omitempty" jsonschema:"Page number"`
}

// ListTagsHandler lists tags in a repository.
func (c *Client) ListTagsHandler(ctx context.Context, req *mcp.CallToolRequest, args ListTagsArgs) (*mcp.CallToolResult, any, error) {
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

	path := fmt.Sprintf("/repositories/%s/%s/refs/tags?pagelen=%d&page=%d",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), pagelen, page)

	result, err := GetPaginated[Tag](c, path)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to list tags: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type CreateTagArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	Name      string `json:"name" jsonschema:"Tag name"`
	Target    string `json:"target" jsonschema:"Target commit hash"`
}

// CreateTagHandler creates a new tag.
func (c *Client) CreateTagHandler(ctx context.Context, req *mcp.CallToolRequest, args CreateTagArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.Name == "" || args.Target == "" {
		return ToolResultError("workspace, repo_slug, name, and target are required"), nil, nil
	}

	body := map[string]interface{}{
		"name":   args.Name,
		"target": map[string]string{"hash": args.Target},
	}

	respData, err := c.Post(fmt.Sprintf("/repositories/%s/%s/refs/tags",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug)), body)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to create tag: %v", err)), nil, nil
	}

	var tag Tag
	if err := json.Unmarshal(respData, &tag); err != nil {
		return ToolResultError(fmt.Sprintf("failed to parse response: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(tag, "", "  ")
	return ToolResultText(string(data)), nil, nil
}
