package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// ListPullRequestsHandler lists pull requests for a repository.
func (c *Client) ListPullRequestsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	state := req.GetString("state", "OPEN")
	pagelen := req.GetInt("pagelen", 25)
	page := req.GetInt("page", 1)
	query := req.GetString("query", "")

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests?state=%s&pagelen=%d&page=%d",
		QueryEscape(workspace), QueryEscape(repoSlug), state, pagelen, page)
	if query != "" {
		path += "&q=" + query
	}

	result, err := GetPaginated[PullRequest](c, path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list pull requests: %v", err)), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// GetPullRequestHandler gets details for a single pull request.
func (c *Client) GetPullRequestHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	pr, err := GetJSON[PullRequest](c, fmt.Sprintf("/repositories/%s/%s/pullrequests/%d",
		QueryEscape(workspace), QueryEscape(repoSlug), prID))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get pull request: %v", err)), nil
	}

	data, _ := json.MarshalIndent(pr, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// CreatePullRequestHandler creates a new pull request.
func (c *Client) CreatePullRequestHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	title, err := req.RequireString("title")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sourceBranch, err := req.RequireString("source_branch")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	body := CreatePRRequest{
		Title: title,
		Source: PREndpoint{
			Branch: &Branch{Name: sourceBranch},
		},
		Description:       req.GetString("description", ""),
		CloseSourceBranch: req.GetBool("close_source_branch", false),
		Draft:             req.GetBool("draft", false),
	}

	destBranch := req.GetString("destination_branch", "")
	if destBranch != "" {
		body.Destination = PREndpoint{
			Branch: &Branch{Name: destBranch},
		}
	}

	respData, err := c.Post(fmt.Sprintf("/repositories/%s/%s/pullrequests",
		QueryEscape(workspace), QueryEscape(repoSlug)), body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create pull request: %v", err)), nil
	}

	var pr PullRequest
	if err := json.Unmarshal(respData, &pr); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse response: %v", err)), nil
	}

	data, _ := json.MarshalIndent(pr, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// UpdatePullRequestHandler updates an existing pull request.
func (c *Client) UpdatePullRequestHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	body := map[string]interface{}{}
	if title := req.GetString("title", ""); title != "" {
		body["title"] = title
	}
	if desc := req.GetString("description", ""); desc != "" {
		body["description"] = desc
	}

	respData, err := c.Put(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d",
		QueryEscape(workspace), QueryEscape(repoSlug), prID), body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update pull request: %v", err)), nil
	}

	var pr PullRequest
	if err := json.Unmarshal(respData, &pr); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse response: %v", err)), nil
	}

	data, _ := json.MarshalIndent(pr, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// MergePullRequestHandler merges a pull request.
func (c *Client) MergePullRequestHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	body := MergePRRequest{
		Type:              "pullrequest",
		CloseSourceBranch: req.GetBool("close_source_branch", false),
		MergeStrategy:     req.GetString("merge_strategy", "merge_commit"),
		Message:           req.GetString("message", ""),
	}

	respData, err := c.Post(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/merge",
		QueryEscape(workspace), QueryEscape(repoSlug), prID), body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to merge pull request: %v", err)), nil
	}

	var pr PullRequest
	if err := json.Unmarshal(respData, &pr); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse response: %v", err)), nil
	}

	data, _ := json.MarshalIndent(pr, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// ApprovePullRequestHandler approves a pull request.
func (c *Client) ApprovePullRequestHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	_, err = c.Post(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/approve",
		QueryEscape(workspace), QueryEscape(repoSlug), prID), nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to approve pull request: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Pull request #%d approved", prID)), nil
}

// UnapprovePullRequestHandler removes approval from a pull request.
func (c *Client) UnapprovePullRequestHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	if err := c.Delete(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/approve",
		QueryEscape(workspace), QueryEscape(repoSlug), prID)); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to unapprove pull request: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Pull request #%d unapproved", prID)), nil
}

// DeclinePullRequestHandler declines a pull request.
func (c *Client) DeclinePullRequestHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	_, err = c.Post(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/decline",
		QueryEscape(workspace), QueryEscape(repoSlug), prID), nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to decline pull request: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Pull request #%d declined", prID)), nil
}

// GetPRDiffHandler gets the diff for a pull request.
func (c *Client) GetPRDiffHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	raw, _, err := c.GetRaw(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/diff",
		QueryEscape(workspace), QueryEscape(repoSlug), prID))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get PR diff: %v", err)), nil
	}

	return mcp.NewToolResultText(string(raw)), nil
}

// GetPRDiffStatHandler gets the diffstat for a pull request.
func (c *Client) GetPRDiffStatHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	result, err := GetPaginated[DiffStat](c, fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/diffstat",
		QueryEscape(workspace), QueryEscape(repoSlug), prID))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get PR diffstat: %v", err)), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// ListPRCommitsHandler lists commits in a pull request.
func (c *Client) ListPRCommitsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	result, err := GetPaginated[Commit](c, fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/commits",
		QueryEscape(workspace), QueryEscape(repoSlug), prID))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list PR commits: %v", err)), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
