package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetFileContentArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	Path      string `json:"path" jsonschema:"Path to the file"`
	Ref       string `json:"ref,omitempty" jsonschema:"Commit hash, branch, or tag (default: HEAD)"`
}

// GetFileContentHandler reads a file's content from the repository.
func (c *Client) GetFileContentHandler(ctx context.Context, req *mcp.CallToolRequest, args GetFileContentArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.Path == "" {
		return ToolResultError("workspace, repo_slug, and path are required"), nil, nil
	}

	var endpoint string
	if args.Ref != "" {
		endpoint = fmt.Sprintf("/repositories/%s/%s/src/%s/%s",
			QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), QueryEscape(args.Ref), args.Path)
	} else {
		endpoint = fmt.Sprintf("/repositories/%s/%s/src/HEAD/%s",
			QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.Path)
	}

	raw, contentType, err := c.GetRaw(endpoint)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to get file content: %v", err)), nil, nil
	}

	// If it looks like JSON (directory listing), format it nicely
	if strings.Contains(contentType, "application/json") {
		var prettyJSON interface{}
		if err := json.Unmarshal(raw, &prettyJSON); err == nil {
			data, _ := json.MarshalIndent(prettyJSON, "", "  ")
			return ToolResultText(string(data)), nil, nil
		}
	}

	return ToolResultText(string(raw)), nil, nil
}

type ListDirectoryArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	Path      string `json:"path,omitempty" jsonschema:"Path to the directory"`
	Ref       string `json:"ref,omitempty" jsonschema:"Commit hash, branch, or tag (default: HEAD)"`
	Pagelen   int    `json:"pagelen,omitempty" jsonschema:"Results per page (default: 100)"`
	MaxDepth  int    `json:"max_depth,omitempty" jsonschema:"Maximum depth of recursion (default: 1)"`
}

// ListDirectoryHandler lists files and directories at a given path.
func (c *Client) ListDirectoryHandler(ctx context.Context, req *mcp.CallToolRequest, args ListDirectoryArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" {
		return ToolResultError("workspace and repo_slug are required"), nil, nil
	}

	pagelen := args.Pagelen
	if pagelen == 0 {
		pagelen = 100
	}
	maxDepth := args.MaxDepth
	if maxDepth == 0 {
		maxDepth = 1
	}

	var endpoint string
	if args.Ref != "" {
		if args.Path != "" {
			endpoint = fmt.Sprintf("/repositories/%s/%s/src/%s/%s",
				QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), QueryEscape(args.Ref), args.Path)
		} else {
			endpoint = fmt.Sprintf("/repositories/%s/%s/src/%s/",
				QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), QueryEscape(args.Ref))
		}
	} else {
		if args.Path != "" {
			endpoint = fmt.Sprintf("/repositories/%s/%s/src/HEAD/%s",
				QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.Path)
		} else {
			endpoint = fmt.Sprintf("/repositories/%s/%s/src/HEAD/",
				QueryEscape(args.Workspace), QueryEscape(args.RepoSlug))
		}
	}

	endpoint += fmt.Sprintf("?pagelen=%d&max_depth=%d", pagelen, maxDepth)

	result, err := GetPaginated[TreeEntry](c, endpoint)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to list directory: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type GetFileHistoryArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	Path      string `json:"path" jsonschema:"Path to the file"`
	Ref       string `json:"ref,omitempty" jsonschema:"Commit hash, branch, or tag (default: HEAD)"`
	Pagelen   int    `json:"pagelen,omitempty" jsonschema:"Results per page (default: 25)"`
}

// GetFileHistoryHandler gets the commit history for a specific file.
func (c *Client) GetFileHistoryHandler(ctx context.Context, req *mcp.CallToolRequest, args GetFileHistoryArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.Path == "" {
		return ToolResultError("workspace, repo_slug, and path are required"), nil, nil
	}

	ref := args.Ref
	if ref == "" {
		ref = "HEAD"
	}
	pagelen := args.Pagelen
	if pagelen == 0 {
		pagelen = 25
	}

	endpoint := fmt.Sprintf("/repositories/%s/%s/filehistory/%s/%s?pagelen=%d",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), QueryEscape(ref), args.Path, pagelen)

	// Filehistory returns commit objects with file metadata
	result, err := GetPaginated[json.RawMessage](c, endpoint)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to get file history: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type SearchCodeArgs struct {
	Workspace   string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug    string `json:"repo_slug" jsonschema:"Repository slug"`
	SearchQuery string `json:"query" jsonschema:"Search query"`
	Pagelen     int    `json:"pagelen,omitempty" jsonschema:"Results per page (default: 25)"`
	Page        int    `json:"page,omitempty" jsonschema:"Page number"`
}

// SearchCodeHandler searches for code in a repository using Bitbucket's code search.
func (c *Client) SearchCodeHandler(ctx context.Context, req *mcp.CallToolRequest, args SearchCodeArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.SearchQuery == "" {
		return ToolResultError("workspace, repo_slug, and query are required"), nil, nil
	}

	pagelen := args.Pagelen
	if pagelen == 0 {
		pagelen = 25
	}
	page := args.Page
	if page == 0 {
		page = 1
	}

	endpoint := fmt.Sprintf("/repositories/%s/%s/search/code?search_query=%s&pagelen=%d&page=%d",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), QueryEscape(args.SearchQuery), pagelen, page)

	raw, err := c.Get(endpoint)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to search code: %v", err)), nil, nil
	}

	var prettyJSON interface{}
	if err := json.Unmarshal(raw, &prettyJSON); err == nil {
		data, _ := json.MarshalIndent(prettyJSON, "", "  ")
		return ToolResultText(string(data)), nil, nil
	}

	return ToolResultText(string(raw)), nil, nil
}
