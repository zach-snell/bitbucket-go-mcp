package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// ListPRCommentsHandler lists comments on a pull request.
func (c *Client) ListPRCommentsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	prID, err := req.RequireInt("pr_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	pagelen := req.GetInt("pagelen", 50)
	page := req.GetInt("page", 1)

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments?pagelen=%d&page=%d",
		QueryEscape(workspace), QueryEscape(repoSlug), prID, pagelen, page)

	result, err := GetPaginated[PRComment](c, path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list PR comments: %v", err)), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// CreatePRCommentHandler creates a comment on a pull request.
func (c *Client) CreatePRCommentHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	prID, err := req.RequireInt("pr_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	content, err := req.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	body := CreateCommentRequest{
		Content: Content{Raw: content},
	}

	// Inline comment support
	filePath := req.GetString("file_path", "")
	if filePath != "" {
		body.Inline = &Inline{
			Path: filePath,
		}
		lineTo := req.GetInt("line_to", 0)
		if lineTo > 0 {
			body.Inline.To = &lineTo
		}
		lineFrom := req.GetInt("line_from", 0)
		if lineFrom > 0 {
			body.Inline.From = &lineFrom
		}
	}

	// Reply to parent comment
	parentID := req.GetInt("parent_id", 0)
	if parentID > 0 {
		body.Parent = &ParentRef{ID: parentID}
	}

	respData, err := c.Post(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments",
		QueryEscape(workspace), QueryEscape(repoSlug), prID), body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create comment: %v", err)), nil
	}

	var comment PRComment
	if err := json.Unmarshal(respData, &comment); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse response: %v", err)), nil
	}

	data, _ := json.MarshalIndent(comment, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// UpdatePRCommentHandler updates an existing comment.
func (c *Client) UpdatePRCommentHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	prID, err := req.RequireInt("pr_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	commentID, err := req.RequireInt("comment_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	content, err := req.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	body := map[string]interface{}{
		"content": map[string]string{"raw": content},
	}

	respData, err := c.Put(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d",
		QueryEscape(workspace), QueryEscape(repoSlug), prID, commentID), body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update comment: %v", err)), nil
	}

	var comment PRComment
	if err := json.Unmarshal(respData, &comment); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse response: %v", err)), nil
	}

	data, _ := json.MarshalIndent(comment, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// DeletePRCommentHandler deletes a comment on a pull request.
func (c *Client) DeletePRCommentHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	prID, err := req.RequireInt("pr_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	commentID, err := req.RequireInt("comment_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := c.Delete(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d",
		QueryEscape(workspace), QueryEscape(repoSlug), prID, commentID)); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to delete comment: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Comment #%d deleted successfully", commentID)), nil
}

// ResolvePRCommentHandler resolves a comment thread.
func (c *Client) ResolvePRCommentHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	prID, err := req.RequireInt("pr_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	commentID, err := req.RequireInt("comment_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	_, err = c.Post(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d/resolve",
		QueryEscape(workspace), QueryEscape(repoSlug), prID, commentID), nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to resolve comment: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Comment #%d resolved", commentID)), nil
}

// UnresolvePRCommentHandler reopens a resolved comment thread.
func (c *Client) UnresolvePRCommentHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	prID, err := req.RequireInt("pr_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	commentID, err := req.RequireInt("comment_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := c.Delete(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d/resolve",
		QueryEscape(workspace), QueryEscape(repoSlug), prID, commentID)); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to unresolve comment: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Comment #%d reopened", commentID)), nil
}
