package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListPullRequestsArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	State     string `json:"state,omitempty" jsonschema:"Filter by state (MERGED, SUPERSEDED, OPEN, DECLINED, default OPEN)"`
	Pagelen   int    `json:"pagelen,omitempty" jsonschema:"Results per page"`
	Page      int    `json:"page,omitempty" jsonschema:"Page number"`
	Query     string `json:"query,omitempty" jsonschema:"Filter query"`
}

// ListPullRequestsHandler lists pull requests for a repository.
func (c *Client) ListPullRequestsHandler(ctx context.Context, req *mcp.CallToolRequest, args ListPullRequestsArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" {
		return ToolResultError("workspace and repo_slug are required"), nil, nil
	}

	state := args.State
	if state == "" {
		state = "OPEN"
	}
	pagelen := args.Pagelen
	if pagelen == 0 {
		pagelen = 25
	}
	page := args.Page
	if page == 0 {
		page = 1
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests?state=%s&pagelen=%d&page=%d",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), state, pagelen, page)
	if args.Query != "" {
		path += "&q=" + QueryEscape(args.Query)
	}

	result, err := GetPaginated[PullRequest](c, path)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to list pull requests: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type GetPullRequestArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	PRID      int    `json:"pr_id" jsonschema:"Pull request ID"`
}

// GetPullRequestHandler gets details for a single pull request.
func (c *Client) GetPullRequestHandler(ctx context.Context, req *mcp.CallToolRequest, args GetPullRequestArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.PRID == 0 {
		return ToolResultError("workspace, repo_slug, and pr_id are required"), nil, nil
	}

	pr, err := GetJSON[PullRequest](c, fmt.Sprintf("/repositories/%s/%s/pullrequests/%d",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.PRID))
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to get pull request: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(pr, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type CreatePullRequestArgs struct {
	Workspace         string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug          string `json:"repo_slug" jsonschema:"Repository slug"`
	Title             string `json:"title" jsonschema:"Title of the pull request"`
	SourceBranch      string `json:"source_branch" jsonschema:"Source branch name"`
	DestinationBranch string `json:"destination_branch,omitempty" jsonschema:"Destination branch name (optional, defaults to repo default)"`
	Description       string `json:"description,omitempty" jsonschema:"Description of the pull request"`
	CloseSourceBranch bool   `json:"close_source_branch,omitempty" jsonschema:"Close source branch on merge"`
	Draft             bool   `json:"draft,omitempty" jsonschema:"Create as a draft PR"`
}

// CreatePullRequestHandler creates a new pull request.
func (c *Client) CreatePullRequestHandler(ctx context.Context, req *mcp.CallToolRequest, args CreatePullRequestArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.Title == "" || args.SourceBranch == "" {
		return ToolResultError("workspace, repo_slug, title, and source_branch are required"), nil, nil
	}

	body := CreatePRRequest{
		Title: args.Title,
		Source: PREndpoint{
			Branch: &Branch{Name: args.SourceBranch},
		},
		Description:       args.Description,
		CloseSourceBranch: args.CloseSourceBranch,
		Draft:             args.Draft,
	}

	if args.DestinationBranch != "" {
		body.Destination = PREndpoint{
			Branch: &Branch{Name: args.DestinationBranch},
		}
	}

	respData, err := c.Post(fmt.Sprintf("/repositories/%s/%s/pullrequests",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug)), body)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to create pull request: %v", err)), nil, nil
	}

	var pr PullRequest
	if err := json.Unmarshal(respData, &pr); err != nil {
		return ToolResultError(fmt.Sprintf("failed to parse response: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(pr, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type UpdatePullRequestArgs struct {
	Workspace   string  `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug    string  `json:"repo_slug" jsonschema:"Repository slug"`
	PRID        int     `json:"pr_id" jsonschema:"Pull request ID"`
	Title       *string `json:"title,omitempty" jsonschema:"New title for the pull request"`
	Description *string `json:"description,omitempty" jsonschema:"New description for the pull request"`
}

// UpdatePullRequestHandler updates an existing pull request.
func (c *Client) UpdatePullRequestHandler(ctx context.Context, req *mcp.CallToolRequest, args UpdatePullRequestArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.PRID == 0 {
		return ToolResultError("workspace, repo_slug, and pr_id are required"), nil, nil
	}

	body := map[string]interface{}{}
	if args.Title != nil {
		body["title"] = *args.Title
	}
	if args.Description != nil {
		body["description"] = *args.Description
	}

	respData, err := c.Put(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.PRID), body)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to update pull request: %v", err)), nil, nil
	}

	var pr PullRequest
	if err := json.Unmarshal(respData, &pr); err != nil {
		return ToolResultError(fmt.Sprintf("failed to parse response: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(pr, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type MergePullRequestArgs struct {
	Workspace         string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug          string `json:"repo_slug" jsonschema:"Repository slug"`
	PRID              int    `json:"pr_id" jsonschema:"Pull request ID"`
	CloseSourceBranch bool   `json:"close_source_branch,omitempty" jsonschema:"Close source branch"`
	MergeStrategy     string `json:"merge_strategy,omitempty" jsonschema:"Merge strategy (e.g. merge_commit, squash, fast_forward, default: merge_commit)"`
	Message           string `json:"message,omitempty" jsonschema:"Commit message"`
}

// MergePullRequestHandler merges a pull request.
func (c *Client) MergePullRequestHandler(ctx context.Context, req *mcp.CallToolRequest, args MergePullRequestArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.PRID == 0 {
		return ToolResultError("workspace, repo_slug, and pr_id are required"), nil, nil
	}

	strategy := args.MergeStrategy
	if strategy == "" {
		strategy = "merge_commit"
	}

	body := MergePRRequest{
		Type:              "pullrequest",
		CloseSourceBranch: args.CloseSourceBranch,
		MergeStrategy:     strategy,
		Message:           args.Message,
	}

	respData, err := c.Post(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/merge",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.PRID), body)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to merge pull request: %v", err)), nil, nil
	}

	var pr PullRequest
	if err := json.Unmarshal(respData, &pr); err != nil {
		return ToolResultError(fmt.Sprintf("failed to parse response: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(pr, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type PullRequestActionArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	PRID      int    `json:"pr_id" jsonschema:"Pull request ID"`
}

// ApprovePullRequestHandler approves a pull request.
//
//nolint:dupl // boilerplate handlers share parameter extraction
func (c *Client) ApprovePullRequestHandler(ctx context.Context, req *mcp.CallToolRequest, args PullRequestActionArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.PRID == 0 {
		return ToolResultError("workspace, repo_slug, and pr_id are required"), nil, nil
	}

	_, err := c.Post(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/approve",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.PRID), nil)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to approve pull request: %v", err)), nil, nil
	}

	return ToolResultText(fmt.Sprintf("Pull request #%d approved", args.PRID)), nil, nil
}

// UnapprovePullRequestHandler removes approval from a pull request.
func (c *Client) UnapprovePullRequestHandler(ctx context.Context, req *mcp.CallToolRequest, args PullRequestActionArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.PRID == 0 {
		return ToolResultError("workspace, repo_slug, and pr_id are required"), nil, nil
	}

	if err := c.Delete(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/approve",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.PRID)); err != nil {
		return ToolResultError(fmt.Sprintf("failed to unapprove pull request: %v", err)), nil, nil
	}

	return ToolResultText(fmt.Sprintf("Pull request #%d unapproved", args.PRID)), nil, nil
}

// DeclinePullRequestHandler declines a pull request.
//
//nolint:dupl // boilerplate handlers share parameter extraction
func (c *Client) DeclinePullRequestHandler(ctx context.Context, req *mcp.CallToolRequest, args PullRequestActionArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.PRID == 0 {
		return ToolResultError("workspace, repo_slug, and pr_id are required"), nil, nil
	}

	_, err := c.Post(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/decline",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.PRID), nil)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to decline pull request: %v", err)), nil, nil
	}

	return ToolResultText(fmt.Sprintf("Pull request #%d declined", args.PRID)), nil, nil
}

// GetPRDiffHandler gets the diff for a pull request.
func (c *Client) GetPRDiffHandler(ctx context.Context, req *mcp.CallToolRequest, args PullRequestActionArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.PRID == 0 {
		return ToolResultError("workspace, repo_slug, and pr_id are required"), nil, nil
	}

	raw, _, err := c.GetRaw(fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/diff",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.PRID))
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to get PR diff: %v", err)), nil, nil
	}

	return ToolResultText(string(raw)), nil, nil
}

// GetPRDiffStatHandler gets the diffstat for a pull request.
func (c *Client) GetPRDiffStatHandler(ctx context.Context, req *mcp.CallToolRequest, args PullRequestActionArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.PRID == 0 {
		return ToolResultError("workspace, repo_slug, and pr_id are required"), nil, nil
	}

	result, err := GetPaginated[DiffStat](c, fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/diffstat",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.PRID))
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to get PR diffstat: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

// ListPRCommitsHandler lists commits in a pull request.
func (c *Client) ListPRCommitsHandler(ctx context.Context, req *mcp.CallToolRequest, args PullRequestActionArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.PRID == 0 {
		return ToolResultError("workspace, repo_slug, and pr_id are required"), nil, nil
	}

	result, err := GetPaginated[Commit](c, fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/commits",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.PRID))
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to list PR commits: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ToolResultText(string(data)), nil, nil
}
