package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListPRCommentsArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	PRID      int    `json:"pr_id" jsonschema:"Pull request ID"`
	Pagelen   int    `json:"pagelen,omitempty" jsonschema:"Results per page (default 50)"`
	Page      int    `json:"page,omitempty" jsonschema:"Page number"`
}

// ListPRCommentsHandler lists comments on a pull request.
func (c *Client) ListPRCommentsHandler(ctx context.Context, req *mcp.CallToolRequest, args ListPRCommentsArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.PRID == 0 {
		return ToolResultError("workspace, repo_slug, and pr_id are required"), nil, nil
	}

	pagelen := args.Pagelen
	if pagelen == 0 {
		pagelen = 50
	}
	page := args.Page
	if page == 0 {
		page = 1
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments?pagelen=%d&page=%d",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.PRID, pagelen, page)

	result, err := GetPaginated[PRComment](c, path)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to list PR comments: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type CreatePRCommentArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	PRID      int    `json:"pr_id" jsonschema:"Pull request ID"`
	Content   string `json:"content" jsonschema:"Markdown content of the comment"`
	FilePath  string `json:"file_path,omitempty" jsonschema:"File path for inline comments"`
	LineTo    int    `json:"line_to,omitempty" jsonschema:"Line number the comment applies to (for new/modified lines)"`
	LineFrom  int    `json:"line_from,omitempty" jsonschema:"Line number the comment applies to (for deleted lines)"`
	ParentID  int    `json:"parent_id,omitempty" jsonschema:"Parent comment ID to reply to"`
}

// CreatePRCommentHandler creates a comment on a pull request.
func (c *Client) CreatePRCommentHandler(ctx context.Context, req *mcp.CallToolRequest, args CreatePRCommentArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.PRID == 0 || args.Content == "" {
		return ToolResultError("workspace, repo_slug, pr_id, and content are required"), nil, nil
	}

	body := CreateCommentRequest{
		Content: Content{Raw: args.Content},
	}

	// Inline comment support
	if args.FilePath != "" {
		body.Inline = &Inline{
			Path: args.FilePath,
		}
		if args.LineTo > 0 {
			lineTo := args.LineTo
			body.Inline.To = &lineTo
		}
		if args.LineFrom > 0 {
			lineFrom := args.LineFrom
			body.Inline.From = &lineFrom
		}
	}

	// Reply to parent comment
	if args.ParentID > 0 {
		body.Parent = &ParentRef{ID: args.ParentID}
	}

	respData, err := c.Post(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.PRID), body)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to create comment: %v", err)), nil, nil
	}

	var comment PRComment
	if err := json.Unmarshal(respData, &comment); err != nil {
		return ToolResultError(fmt.Sprintf("failed to parse response: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(comment, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type UpdatePRCommentArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	PRID      int    `json:"pr_id" jsonschema:"Pull request ID"`
	CommentID int    `json:"comment_id" jsonschema:"Comment ID to update"`
	Content   string `json:"content" jsonschema:"New markdown content"`
}

// UpdatePRCommentHandler updates an existing comment.
func (c *Client) UpdatePRCommentHandler(ctx context.Context, req *mcp.CallToolRequest, args UpdatePRCommentArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.PRID == 0 || args.CommentID == 0 || args.Content == "" {
		return ToolResultError("workspace, repo_slug, pr_id, comment_id, and content are required"), nil, nil
	}

	body := map[string]interface{}{
		"content": map[string]string{"raw": args.Content},
	}

	respData, err := c.Put(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.PRID, args.CommentID), body)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to update comment: %v", err)), nil, nil
	}

	var comment PRComment
	if err := json.Unmarshal(respData, &comment); err != nil {
		return ToolResultError(fmt.Sprintf("failed to parse response: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(comment, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type CommentActionArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	PRID      int    `json:"pr_id" jsonschema:"Pull request ID"`
	CommentID int    `json:"comment_id" jsonschema:"Comment ID"`
}

// DeletePRCommentHandler deletes a comment on a pull request.
//
//nolint:dupl // boilerplate handlers share parameter extraction
func (c *Client) DeletePRCommentHandler(ctx context.Context, req *mcp.CallToolRequest, args CommentActionArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.PRID == 0 || args.CommentID == 0 {
		return ToolResultError("workspace, repo_slug, pr_id, and comment_id are required"), nil, nil
	}

	if err := c.Delete(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.PRID, args.CommentID)); err != nil {
		return ToolResultError(fmt.Sprintf("failed to delete comment: %v", err)), nil, nil
	}

	return ToolResultText(fmt.Sprintf("Comment #%d deleted successfully", args.CommentID)), nil, nil
}

// ResolvePRCommentHandler resolves a comment thread.
func (c *Client) ResolvePRCommentHandler(ctx context.Context, req *mcp.CallToolRequest, args CommentActionArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.PRID == 0 || args.CommentID == 0 {
		return ToolResultError("workspace, repo_slug, pr_id, and comment_id are required"), nil, nil
	}

	_, err := c.Post(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d/resolve",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.PRID, args.CommentID), nil)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to resolve comment: %v", err)), nil, nil
	}

	return ToolResultText(fmt.Sprintf("Comment #%d resolved", args.CommentID)), nil, nil
}

// UnresolvePRCommentHandler reopens a resolved comment thread.
//
//nolint:dupl // boilerplate handlers share parameter extraction
func (c *Client) UnresolvePRCommentHandler(ctx context.Context, req *mcp.CallToolRequest, args CommentActionArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.PRID == 0 || args.CommentID == 0 {
		return ToolResultError("workspace, repo_slug, pr_id, and comment_id are required"), nil, nil
	}

	if err := c.Delete(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d/resolve",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.PRID, args.CommentID)); err != nil {
		return ToolResultError(fmt.Sprintf("failed to unresolve comment: %v", err)), nil, nil
	}

	return ToolResultText(fmt.Sprintf("Comment #%d reopened", args.CommentID)), nil, nil
}
