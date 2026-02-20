package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zach-snell/bitbucket-go-mcp/internal/bitbucket"
	mcpserver "github.com/zach-snell/bitbucket-go-mcp/internal/server"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "auth", "login":
			runAuth()
			return
		case "status":
			runStatus()
			return
		case "logout":
			runLogout()
			return
		case "help", "--help", "-h":
			printUsage()
			return
		}
	}

	runServer()
}

func runServer() {
	port := flag.Int("port", 0, "Port to listen on for HTTP Streamable transport")
	_ = flag.CommandLine.Parse(os.Args[1:])

	// Priority: env vars > stored credentials
	username := os.Getenv("BITBUCKET_USERNAME")
	password := os.Getenv("BITBUCKET_APP_PASSWORD")
	token := os.Getenv("BITBUCKET_ACCESS_TOKEN")

	var s *mcp.Server

	if token != "" || (username != "" && password != "") {
		s = mcpserver.New(username, password, token)
	} else {
		creds, err := bitbucket.LoadCredentials()
		if err != nil {
			fmt.Fprintf(os.Stderr, "No credentials found. Either:\n")
			fmt.Fprintf(os.Stderr, "  1. Run: bitbucket-mcp auth          (API token — recommended)\n")
			fmt.Fprintf(os.Stderr, "  2. Run: bitbucket-mcp auth --oauth   (OAuth via browser)\n")
			fmt.Fprintf(os.Stderr, "  3. Set BITBUCKET_ACCESS_TOKEN env var\n")
			fmt.Fprintf(os.Stderr, "  4. Set BITBUCKET_USERNAME + BITBUCKET_APP_PASSWORD env vars\n")
			os.Exit(1)
		}

		switch {
		case creds.IsAPIToken():
			s = mcpserver.New(creds.Email, creds.APIToken, "")
		case creds.IsOAuth():
			s = mcpserver.NewFromOAuth(creds)
		default:
			fmt.Fprintf(os.Stderr, "Unknown auth type in stored credentials: %s\n", creds.AuthType)
			os.Exit(1)
		}
	}

	if *port != 0 {
		fmt.Printf("Starting Bitbucket MCP Server on :%d (HTTP Streamable)\n", *port)
		handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
			return s
		}, &mcp.StreamableHTTPOptions{JSONResponse: false})

		srv := &http.Server{
			Addr:              fmt.Sprintf(":%d", *port),
			Handler:           handler,
			ReadHeaderTimeout: 3 * time.Second,
		}

		if err := srv.ListenAndServe(); err != nil {
			fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}
}

func runAuth() {
	// Check for --oauth flag
	if slices.Contains(os.Args[2:], "--oauth") {
		runOAuthLogin()
		return
	}

	// Default: API Token flow
	if err := bitbucket.APITokenLogin(); err != nil {
		fmt.Fprintf(os.Stderr, "auth failed: %v\n", err)
		os.Exit(1)
	}
}

func runOAuthLogin() {
	clientID := os.Getenv("BITBUCKET_OAUTH_CLIENT_ID")
	clientSecret := os.Getenv("BITBUCKET_OAUTH_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		fmt.Fprintf(os.Stderr, "OAuth credentials required. Set:\n")
		fmt.Fprintf(os.Stderr, "  BITBUCKET_OAUTH_CLIENT_ID\n")
		fmt.Fprintf(os.Stderr, "  BITBUCKET_OAUTH_CLIENT_SECRET\n\n")
		fmt.Fprintf(os.Stderr, "Create an OAuth consumer at:\n")
		fmt.Fprintf(os.Stderr, "  Bitbucket > Workspace Settings > OAuth consumers > Add consumer\n")
		fmt.Fprintf(os.Stderr, "  Callback URL: http://localhost:<any-port>/callback\n")
		fmt.Fprintf(os.Stderr, "  Scopes: repository, repository:write, pullrequest, pullrequest:write,\n")
		fmt.Fprintf(os.Stderr, "          pipeline, pipeline:write, account\n")
		os.Exit(1)
	}

	if err := bitbucket.OAuthLogin(clientID, clientSecret); err != nil {
		fmt.Fprintf(os.Stderr, "auth failed: %v\n", err)
		os.Exit(1)
	}
}

func runStatus() {
	creds, err := bitbucket.LoadCredentials()
	if err != nil {
		if os.Getenv("BITBUCKET_ACCESS_TOKEN") != "" {
			fmt.Println("Authenticated via BITBUCKET_ACCESS_TOKEN environment variable")
			return
		}
		if os.Getenv("BITBUCKET_USERNAME") != "" && os.Getenv("BITBUCKET_APP_PASSWORD") != "" {
			fmt.Println("Authenticated via BITBUCKET_USERNAME + BITBUCKET_APP_PASSWORD environment variables")
			return
		}
		fmt.Println("Not authenticated. Run: bitbucket-mcp auth")
		return
	}

	path, _ := bitbucket.CredentialsPath()

	switch {
	case creds.IsAPIToken():
		fmt.Println("Authenticated via API Token (Basic Auth)")
		fmt.Printf("  Email:   %s\n", creds.Email)
		if len(creds.APIToken) > 8 {
			fmt.Printf("  Token:   %s...%s\n", creds.APIToken[:4], creds.APIToken[len(creds.APIToken)-4:])
		} else {
			fmt.Println("  Token:   ****")
		}
		fmt.Printf("  Stored:  %s\n", creds.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  File:    %s\n", path)

	case creds.IsOAuth():
		fmt.Println("Authenticated via OAuth 2.0 (Bearer Auth)")
		fmt.Printf("  Scopes:  %s\n", creds.Scopes)
		fmt.Printf("  Stored:  %s\n", creds.CreatedAt.Format("2006-01-02 15:04:05"))
		if creds.IsExpired() {
			fmt.Println("  Status:  expired (will auto-refresh)")
		} else {
			fmt.Println("  Status:  valid")
		}
		fmt.Printf("  File:    %s\n", path)
	}
}

func runLogout() {
	if err := bitbucket.RemoveCredentials(); err != nil {
		fmt.Fprintf(os.Stderr, "error removing credentials: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Logged out. Credentials removed.")
}

func printUsage() {
	fmt.Println(`bitbucket-mcp - Bitbucket Cloud MCP server

Usage:
  bitbucket-mcp               Start MCP server (stdio transport)
  bitbucket-mcp --port 8080   Start MCP server (HTTP Streamable transport)
  bitbucket-mcp auth          Authenticate with API token (recommended)
  bitbucket-mcp auth --oauth  Authenticate via OAuth (opens browser)
  bitbucket-mcp status        Show current auth status
  bitbucket-mcp logout        Remove stored credentials
  bitbucket-mcp help          Show this help

Authentication (in priority order):
  1. BITBUCKET_ACCESS_TOKEN env var (Bearer token)
  2. BITBUCKET_USERNAME + BITBUCKET_APP_PASSWORD env vars (Basic Auth)
  3. Stored credentials from 'bitbucket-mcp auth'

API Token setup (recommended):
  1. Go to https://bitbucket.org/account/settings/api-tokens/
  2. Create a token with needed scopes
  3. Run: bitbucket-mcp auth

OAuth setup (requires workspace admin):
  1. Create an OAuth consumer in Bitbucket workspace settings
  2. Set BITBUCKET_OAUTH_CLIENT_ID and BITBUCKET_OAUTH_CLIENT_SECRET
  3. Run: bitbucket-mcp auth --oauth`)
}
