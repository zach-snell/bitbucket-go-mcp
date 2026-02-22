# Contributing to bbkt

First off, thank you for considering contributing to `bbkt`! It's people like you that make open source tools great.

This document serves as a set of guidelines for contributing to this project. 

## Getting Started

`bbkt` is built using Go. You will need **Go 1.25** or higher installed on your machine.

1. **Fork the repository** on GitHub.
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/bbkt.git
   cd bbkt
   ```
3. **Install Dependencies:**
   ```bash
   go mod download
   ```
4. **Create a branch** for your feature or bugfix:
   ```bash
   git checkout -b feature/my-new-feature
   ```

## Development Workflow

The project is broken into two primary interfaces:
1. **The CLI (`cmd/cli`)**: The interactive, standard terminal interface.
2. **The MCP Server (`internal/mcp`)**: The Model Context Protocol implementation exposing tools to AI Agents.

Both interfaces rely on the core API client found in `internal/bitbucket`. 

### Making Changes
If you introduce a new feature, try to expose it in *both* the CLI and the MCP Server. 

- **Adding a CLI Command**: Scaffold the command in `cmd/cli/` using the Cobra framework. Refer to existing commands for examples on how to invoke the UI prompt wizards for missing arguments.
- **Adding an MCP Tool**: We heavily multiplex our tools. Avoid creating entirely new top-level tools unless absolutely necessary. Instead, add a new `action` argument to an existing tool (e.g., add `action: "revert"` to `manage_pull_requests`) inside `internal/mcp/tools.go`.

### Testing Your Changes

Before submitting a Pull Request, ensure your code passes both the compiler and the linter. 

1. **Run the Go Formatter**:
   ```bash
   gofumpt -w .
   ```
2. **Run all tests**:
   ```bash
   go test -race ./...
   ```
   *(Note: As of v0.1.1, automated test coverage is sparse. We highly encourage adding `_test.go` files alongside any new code you contribute!)*
3. **Run the Linter**:
   We use `golangci-lint` to enforce code quality. Make sure it passes locally:
   ```bash
   golangci-lint run ./...
   ```

## Pull Request Process

1. Ensure your PR follows the `.github/PULL_REQUEST_TEMPLATE.md` guidelines.
2. Write a clear, concise title and description. Link any related issues using `Fixes #000`.
3. Ensure the GitHub Actions CI pipeline passes successfully. We enforce strict linting.
4. Once approved by a maintainer, your PR will be squash-merged into `main`.

## Code of Conduct

By participating in this project, you are expected to uphold standard professional conduct. Be respectful to other contributors and users, accept constructive feedback gracefully, and focus on collaborative improvement. 
