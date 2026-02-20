# Bitbucket Go MCP Server

A Model Context Protocol (MCP) server written in Go that provides programmatic integration with Bitbucket workspaces and repositories.

## Features

- **Read Capabilities**: View branches, commits, PRs, pipelines, diffs, and repositories.
- **Write Capabilities**: Create/merge PRs, add PR comments, resolve PR threads, and trigger pipelines.
- **Authentication**: Supports standard `app_password` credentials or an interactive OAuth 2.0 flow if no credentials are provided.

## Installation

### From Source
```bash
# Clone the repository
git clone https://github.com/your-username/bitbucket-go-mcp.git
cd bitbucket-go-mcp

# Build the executable
go build -o bitbucket-mcp ./cmd/server
```

### From GitHub Releases
Download the appropriate binary for your system (Linux, macOS, Windows) from the [Releases](https://github.com/your-username/bitbucket-go-mcp/releases) page.

## Configuration & Usage

If you intend to use this with an MCP client (such as Claude Desktop or Cursor), add it to your client's configuration file as a local command:

```json
{
  "mcpServers": {
    "bitbucket": {
      "command": "/absolute/path/to/bitbucket-mcp",
      "env": {
        "BITBUCKET_USERNAME": "your-username",
        "BITBUCKET_APP_PASSWORD": "your-app-password"
      }
    }
  }
}
```

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `BITBUCKET_USERNAME` | Your Bitbucket username | No (but recommended for App Passwords) |
| `BITBUCKET_APP_PASSWORD` | An app password with api access | No (If omitted, triggers OAuth 2.0 browser flow) |
| `BITBUCKET_CLIENT_ID` | OAuth 2.0 Client ID | Only if using OAuth |
| `BITBUCKET_CLIENT_SECRET` | OAuth 2.0 Client Secret | Only if using OAuth |

## Tools Provided

- `list_repositories`: List repositories in a workspace.
- `list_pull_requests`: List pull requests for a repository.
- `create_pull_request`: Open a new pull request.
- `merge_pull_request`: Merge an existing pull request.
- `approve_pull_request`: Approve a pull request.
- `create_pr_comment`: Reply to or add an inline comment on a pull request.
- `list_pr_commits`: See exactly which commits are included in a PR.
- `list_pipelines`: View build/deployment pipelines across a repository.
- `get_pipeline`: Fetch specific details about a pipeline run.
- `list_branches`: View existing branches.
- `list_commits`: View recent commits to a repository.
- `get_diffstat`: See file additions and deletions for a specific commit or PR.

## Development

Requirements:
- Go 1.25+

```bash
# Run tests
go test ./...

# Run the linter
golangci-lint run ./...
```
