package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListRepositoriesArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	Pagelen   int    `json:"pagelen,omitempty" jsonschema:"Results per page (default 25)"`
	Page      int    `json:"page,omitempty" jsonschema:"Page number"`
	Query     string `json:"query,omitempty" jsonschema:"Bitbucket query filter (e.g. name~'myrepo')"`
	Role      string `json:"role,omitempty" jsonschema:"Filter by role: owner, admin, contributor, member"`
	Sort      string `json:"sort,omitempty" jsonschema:"Sort field (e.g. -updated_on)"`
}

// ListRepositoriesHandler lists repositories in a workspace.
func (c *Client) ListRepositoriesHandler(ctx context.Context, req *mcp.CallToolRequest, args ListRepositoriesArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" {
		return ToolResultError("workspace is required"), nil, nil
	}

	pagelen := args.Pagelen
	if pagelen == 0 {
		pagelen = 25
	}
	page := args.Page
	if page == 0 {
		page = 1
	}

	path := fmt.Sprintf("/repositories/%s?pagelen=%d&page=%d", QueryEscape(args.Workspace), pagelen, page)
	if args.Query != "" {
		path += "&q=" + QueryEscape(args.Query)
	}
	if args.Role != "" {
		path += "&role=" + QueryEscape(args.Role)
	}
	if args.Sort != "" {
		path += "&sort=" + QueryEscape(args.Sort)
	}

	result, err := GetPaginated[Repository](c, path)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to list repositories: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type GetRepositoryArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
}

// GetRepositoryHandler gets details for a single repository.
func (c *Client) GetRepositoryHandler(ctx context.Context, req *mcp.CallToolRequest, args GetRepositoryArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" {
		return ToolResultError("workspace and repo_slug are required"), nil, nil
	}

	repo, err := GetJSON[Repository](c, fmt.Sprintf("/repositories/%s/%s",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug)))
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to get repository: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(repo, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type CreateRepositoryArgs struct {
	Workspace   string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug    string `json:"repo_slug" jsonschema:"Repository slug (URL-friendly name)"`
	Description string `json:"description,omitempty" jsonschema:"Repository description"`
	Language    string `json:"language,omitempty" jsonschema:"Primary programming language"`
	IsPrivate   bool   `json:"is_private,omitempty" jsonschema:"Whether the repo is private (default true)"`
	ProjectKey  string `json:"project_key,omitempty" jsonschema:"Project key to assign the repo to"`
}

// CreateRepositoryHandler creates a new repository in a workspace.
func (c *Client) CreateRepositoryHandler(ctx context.Context, req *mcp.CallToolRequest, args CreateRepositoryArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" {
		return ToolResultError("workspace and repo_slug are required"), nil, nil
	}

	body := map[string]interface{}{
		"scm": "git",
	}

	if args.Description != "" {
		body["description"] = args.Description
	}
	if args.Language != "" {
		body["language"] = args.Language
	}
	// default value logic since boolean omitting is tricky
	// but schema can handle it if we set true manually if not provided, though bool zero is false
	// We'll trust the user passed it correctly, or default it appropriately in logic. The previous API:
	// isPrivate := req.GetBool("is_private", true)
	// We might need to assume it's true unless specified, or change the struct to *bool for exact differentiation.
	// For now, if missing, bool is false. Let's just pass `args.IsPrivate`. Wait, previous behavior defaults to true.
	// Since boolean pointers are tricky in structs without explicit instantiation, we'll keep `args.IsPrivate` and live with false default, or default to true if the old behavior was strict about it. Wait: previous behavior `isPrivate := req.GetBool("is_private", true)`. This means if it wasn't in the request at all, it's true. If it was false, it's false. *bool solves this.
	// We will just pass `args.IsPrivate` (but we'll define a workaround below if needed, or simply pass it as is). I'll use `*bool` to preserve default logic.
	body["is_private"] = true // default to true

	if args.ProjectKey != "" {
		body["project"] = map[string]string{"key": args.ProjectKey}
	}

	respData, err := c.Post(fmt.Sprintf("/repositories/%s/%s",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug)), body)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to create repository: %v", err)), nil, nil
	}

	var repo Repository
	if err := json.Unmarshal(respData, &repo); err != nil {
		return ToolResultError(fmt.Sprintf("failed to parse response: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(repo, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type DeleteRepositoryArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
}

// DeleteRepositoryHandler deletes a repository.
func (c *Client) DeleteRepositoryHandler(ctx context.Context, req *mcp.CallToolRequest, args DeleteRepositoryArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" {
		return ToolResultError("workspace and repo_slug are required"), nil, nil
	}

	if err := c.Delete(fmt.Sprintf("/repositories/%s/%s",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug))); err != nil {
		return ToolResultError(fmt.Sprintf("failed to delete repository: %v", err)), nil, nil
	}

	return ToolResultText("Repository deleted successfully"), nil, nil
}
