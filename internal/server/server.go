package server

import (
	"context"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zach-snell/bitbucket-go-mcp/internal/bitbucket"
)

// New creates and configures the Bitbucket MCP server with all tools registered.
func New(username, password, token string) *mcp.Server {
	client := bitbucket.NewClient(username, password, token)
	return newServer(client)
}

// NewFromOAuth creates the MCP server from stored OAuth credentials with auto-refresh.
func NewFromOAuth(creds *bitbucket.Credentials) *mcp.Server {
	client := bitbucket.NewClientFromOAuth(creds)
	return newServer(client)
}

func newServer(client *bitbucket.Client) *mcp.Server {
	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "bitbucket-mcp",
			Version: "1.0.0",
		},
		nil,
	)

	registerTools(s, client)
	return s
}

// addTool is a helper function to conditionally register a generic tool handler
func addTool[In any](s *mcp.Server, disabled map[string]bool, tool mcp.Tool, handler func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, any, error)) {
	if disabled[tool.Name] {
		return
	}
	mcp.AddTool(s, &tool, handler)
}

func registerTools(s *mcp.Server, c *bitbucket.Client) {
	disabledToolsEnv := os.Getenv("BITBUCKET_DISABLED_TOOLS")
	disabled := make(map[string]bool)
	if disabledToolsEnv != "" {
		for _, t := range strings.Split(disabledToolsEnv, ",") {
			disabled[strings.TrimSpace(t)] = true
		}
	}

	// ─── Workspaces ──────────────────────────────────────────────────
	addTool(s, disabled, mcp.Tool{
		Name:        "list_workspaces",
		Description: "List Bitbucket workspaces accessible to the authenticated user",
	}, c.ListWorkspacesHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "get_workspace",
		Description: "Get details for a Bitbucket workspace",
	}, c.GetWorkspaceHandler)

	// ─── Repositories ────────────────────────────────────────────────
	addTool(s, disabled, mcp.Tool{
		Name:        "list_repositories",
		Description: "List repositories in a Bitbucket workspace",
	}, c.ListRepositoriesHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "get_repository",
		Description: "Get details for a specific repository",
	}, c.GetRepositoryHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "create_repository",
		Description: "Create a new repository in a workspace",
	}, c.CreateRepositoryHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "delete_repository",
		Description: "Delete a repository (DESTRUCTIVE - cannot be undone)",
	}, c.DeleteRepositoryHandler)

	// ─── Branches ────────────────────────────────────────────────────
	addTool(s, disabled, mcp.Tool{
		Name:        "list_branches",
		Description: "List branches in a repository",
	}, c.ListBranchesHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "create_branch",
		Description: "Create a new branch from a commit hash",
	}, c.CreateBranchHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "delete_branch",
		Description: "Delete a branch",
	}, c.DeleteBranchHandler)

	// ─── Tags ────────────────────────────────────────────────────────
	addTool(s, disabled, mcp.Tool{
		Name:        "list_tags",
		Description: "List tags in a repository",
	}, c.ListTagsHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "create_tag",
		Description: "Create a new tag at a specific commit",
	}, c.CreateTagHandler)

	// ─── Commits ─────────────────────────────────────────────────────
	addTool(s, disabled, mcp.Tool{
		Name:        "list_commits",
		Description: "List commits in a repository, optionally filtered by branch/revision",
	}, c.ListCommitsHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "get_commit",
		Description: "Get details for a single commit",
	}, c.GetCommitHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "get_diff",
		Description: "Get diff for a commit or between two revisions (e.g. 'hash1..hash2')",
	}, c.GetDiffHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "get_diffstat",
		Description: "Get diff statistics (files changed, lines added/removed)",
	}, c.GetDiffStatHandler)

	// ─── Pull Requests ───────────────────────────────────────────────
	addTool(s, disabled, mcp.Tool{
		Name:        "list_pull_requests",
		Description: "List pull requests for a repository",
	}, c.ListPullRequestsHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "get_pull_request",
		Description: "Get details for a specific pull request",
	}, c.GetPullRequestHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "create_pull_request",
		Description: "Create a new pull request",
	}, c.CreatePullRequestHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "update_pull_request",
		Description: "Update a pull request's title or description",
	}, c.UpdatePullRequestHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "merge_pull_request",
		Description: "Merge a pull request",
	}, c.MergePullRequestHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "approve_pull_request",
		Description: "Approve a pull request",
	}, c.ApprovePullRequestHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "unapprove_pull_request",
		Description: "Remove approval from a pull request",
	}, c.UnapprovePullRequestHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "decline_pull_request",
		Description: "Decline a pull request",
	}, c.DeclinePullRequestHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "get_pr_diff",
		Description: "Get the diff for a pull request",
	}, c.GetPRDiffHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "get_pr_diffstat",
		Description: "Get diff statistics for a pull request (files changed, lines added/removed)",
	}, c.GetPRDiffStatHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "list_pr_commits",
		Description: "List commits in a pull request",
	}, c.ListPRCommitsHandler)

	// ─── PR Comments ─────────────────────────────────────────────────
	addTool(s, disabled, mcp.Tool{
		Name:        "list_pr_comments",
		Description: "List comments on a pull request",
	}, c.ListPRCommentsHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "create_pr_comment",
		Description: "Add a comment to a pull request. Supports inline comments on specific files/lines and replies to existing comments.",
	}, c.CreatePRCommentHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "update_pr_comment",
		Description: "Update an existing comment on a pull request",
	}, c.UpdatePRCommentHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "delete_pr_comment",
		Description: "Delete a comment from a pull request",
	}, c.DeletePRCommentHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "resolve_pr_comment",
		Description: "Resolve a comment thread on a pull request",
	}, c.ResolvePRCommentHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "unresolve_pr_comment",
		Description: "Reopen a resolved comment thread",
	}, c.UnresolvePRCommentHandler)

	// ─── Source / File Browsing ──────────────────────────────────────
	addTool(s, disabled, mcp.Tool{
		Name:        "get_file_content",
		Description: "Read a file's content from the repository at a given revision",
	}, c.GetFileContentHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "list_directory",
		Description: "List files and directories at a path in the repository",
	}, c.ListDirectoryHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "get_file_history",
		Description: "Get the commit history for a specific file",
	}, c.GetFileHistoryHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "search_code",
		Description: "Search for code in a repository",
	}, c.SearchCodeHandler)

	// ─── Pipelines ───────────────────────────────────────────────────
	addTool(s, disabled, mcp.Tool{
		Name:        "list_pipelines",
		Description: "List pipeline runs for a repository",
	}, c.ListPipelinesHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "get_pipeline",
		Description: "Get details for a specific pipeline run",
	}, c.GetPipelineHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "trigger_pipeline",
		Description: "Trigger a new pipeline run on a branch",
	}, c.TriggerPipelineHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "stop_pipeline",
		Description: "Stop a running pipeline",
	}, c.StopPipelineHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "list_pipeline_steps",
		Description: "List steps in a pipeline run",
	}, c.ListPipelineStepsHandler)

	addTool(s, disabled, mcp.Tool{
		Name:        "get_pipeline_step_log",
		Description: "Get the log output for a pipeline step",
	}, c.GetPipelineStepLogHandler)
}
