package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// GetFileContentHandler reads a file's content from the repository.
func (c *Client) GetFileContentHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	filePath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	ref := req.GetString("ref", "")

	var endpoint string
	if ref != "" {
		endpoint = fmt.Sprintf("/repositories/%s/%s/src/%s/%s",
			QueryEscape(workspace), QueryEscape(repoSlug), QueryEscape(ref), filePath)
	} else {
		endpoint = fmt.Sprintf("/repositories/%s/%s/src/HEAD/%s",
			QueryEscape(workspace), QueryEscape(repoSlug), filePath)
	}

	raw, contentType, err := c.GetRaw(endpoint)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get file content: %v", err)), nil
	}

	// If it looks like JSON (directory listing), format it nicely
	if strings.Contains(contentType, "application/json") {
		var prettyJSON interface{}
		if err := json.Unmarshal(raw, &prettyJSON); err == nil {
			data, _ := json.MarshalIndent(prettyJSON, "", "  ")
			return mcp.NewToolResultText(string(data)), nil
		}
	}

	return mcp.NewToolResultText(string(raw)), nil
}

// ListDirectoryHandler lists files and directories at a given path.
func (c *Client) ListDirectoryHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	dirPath := req.GetString("path", "")
	ref := req.GetString("ref", "")
	pagelen := req.GetInt("pagelen", 100)
	maxDepth := req.GetInt("max_depth", 1)

	var endpoint string
	if ref != "" {
		if dirPath != "" {
			endpoint = fmt.Sprintf("/repositories/%s/%s/src/%s/%s",
				QueryEscape(workspace), QueryEscape(repoSlug), QueryEscape(ref), dirPath)
		} else {
			endpoint = fmt.Sprintf("/repositories/%s/%s/src/%s/",
				QueryEscape(workspace), QueryEscape(repoSlug), QueryEscape(ref))
		}
	} else {
		if dirPath != "" {
			endpoint = fmt.Sprintf("/repositories/%s/%s/src/HEAD/%s",
				QueryEscape(workspace), QueryEscape(repoSlug), dirPath)
		} else {
			endpoint = fmt.Sprintf("/repositories/%s/%s/src/HEAD/",
				QueryEscape(workspace), QueryEscape(repoSlug))
		}
	}

	endpoint += fmt.Sprintf("?pagelen=%d&max_depth=%d", pagelen, maxDepth)

	result, err := GetPaginated[TreeEntry](c, endpoint)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list directory: %v", err)), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// GetFileHistoryHandler gets the commit history for a specific file.
func (c *Client) GetFileHistoryHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	filePath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	ref := req.GetString("ref", "HEAD")
	pagelen := req.GetInt("pagelen", 25)

	endpoint := fmt.Sprintf("/repositories/%s/%s/filehistory/%s/%s?pagelen=%d",
		QueryEscape(workspace), QueryEscape(repoSlug), QueryEscape(ref), filePath, pagelen)

	// Filehistory returns commit objects with file metadata
	result, err := GetPaginated[json.RawMessage](c, endpoint)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get file history: %v", err)), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// SearchCodeHandler searches for code in a repository using Bitbucket's code search.
func (c *Client) SearchCodeHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	searchQuery, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	pagelen := req.GetInt("pagelen", 25)
	page := req.GetInt("page", 1)

	endpoint := fmt.Sprintf("/repositories/%s/%s/search/code?search_query=%s&pagelen=%d&page=%d",
		QueryEscape(workspace), QueryEscape(repoSlug), QueryEscape(searchQuery), pagelen, page)

	raw, err := c.Get(endpoint)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to search code: %v", err)), nil
	}

	var prettyJSON interface{}
	if err := json.Unmarshal(raw, &prettyJSON); err == nil {
		data, _ := json.MarshalIndent(prettyJSON, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}

	return mcp.NewToolResultText(string(raw)), nil
}
