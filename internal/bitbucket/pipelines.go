package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListPipelinesArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	Pagelen   int    `json:"pagelen,omitempty" jsonschema:"Results per page"`
	Page      int    `json:"page,omitempty" jsonschema:"Page number"`
	Sort      string `json:"sort,omitempty" jsonschema:"Sort field (default -created_on)"`
	Status    string `json:"status,omitempty" jsonschema:"Filter by status"`
}

// ListPipelinesHandler lists pipeline runs for a repository.
func (c *Client) ListPipelinesHandler(ctx context.Context, req *mcp.CallToolRequest, args ListPipelinesArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" {
		return ToolResultError("workspace and repo_slug are required"), nil, nil
	}

	pagelen := args.Pagelen
	if pagelen == 0 {
		pagelen = 25
	}
	page := args.Page
	if page == 0 {
		page = 1
	}
	sort := args.Sort
	if sort == "" {
		sort = "-created_on"
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines?pagelen=%d&page=%d&sort=%s",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), pagelen, page, QueryEscape(sort))

	if args.Status != "" {
		path += "&status=" + QueryEscape(args.Status)
	}

	result, err := GetPaginated[Pipeline](c, path)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to list pipelines: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type GetPipelineArgs struct {
	Workspace    string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug     string `json:"repo_slug" jsonschema:"Repository slug"`
	PipelineUUID string `json:"pipeline_uuid" jsonschema:"Pipeline UUID"`
}

// GetPipelineHandler gets details for a single pipeline run.
func (c *Client) GetPipelineHandler(ctx context.Context, req *mcp.CallToolRequest, args GetPipelineArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.PipelineUUID == "" {
		return ToolResultError("workspace, repo_slug, and pipeline_uuid are required"), nil, nil
	}

	pipe, err := GetJSON[Pipeline](c, fmt.Sprintf("/repositories/%s/%s/pipelines/%s",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.PipelineUUID))
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to get pipeline: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(pipe, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type TriggerPipelineArgs struct {
	Workspace string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug  string `json:"repo_slug" jsonschema:"Repository slug"`
	RefName   string `json:"ref_name" jsonschema:"Branch or tag name to run pipeline on"`
	RefType   string `json:"ref_type,omitempty" jsonschema:"Reference type: branch or tag (default branch)"`
	Pattern   string `json:"pattern,omitempty" jsonschema:"Custom pipeline pattern name to trigger"`
}

// TriggerPipelineHandler triggers a new pipeline run.
func (c *Client) TriggerPipelineHandler(ctx context.Context, req *mcp.CallToolRequest, args TriggerPipelineArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.RefName == "" {
		return ToolResultError("workspace, repo_slug, and ref_name are required"), nil, nil
	}

	refType := args.RefType
	if refType == "" {
		refType = "branch"
	}

	body := TriggerPipelineRequest{
		Target: PipeTriggerTarget{
			Type:    "pipeline_ref_target",
			RefType: refType,
			RefName: args.RefName,
		},
	}

	if args.Pattern != "" {
		body.Target.Selector = &PipelineSelector{
			Type:    "custom",
			Pattern: args.Pattern,
		}
	}

	respData, err := c.Post(fmt.Sprintf("/repositories/%s/%s/pipelines",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug)), body)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to trigger pipeline: %v", err)), nil, nil
	}

	var pipe Pipeline
	if err := json.Unmarshal(respData, &pipe); err != nil {
		return ToolResultError(fmt.Sprintf("failed to parse response: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(pipe, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type StopPipelineArgs struct {
	Workspace    string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug     string `json:"repo_slug" jsonschema:"Repository slug"`
	PipelineUUID string `json:"pipeline_uuid" jsonschema:"Pipeline UUID to stop"`
}

// StopPipelineHandler stops a running pipeline.
func (c *Client) StopPipelineHandler(ctx context.Context, req *mcp.CallToolRequest, args StopPipelineArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.PipelineUUID == "" {
		return ToolResultError("workspace, repo_slug, and pipeline_uuid are required"), nil, nil
	}

	_, err := c.Post(fmt.Sprintf("/repositories/%s/%s/pipelines/%s/stopPipeline",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.PipelineUUID), nil)
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to stop pipeline: %v", err)), nil, nil
	}

	return ToolResultText("Pipeline stopped successfully"), nil, nil
}

type ListPipelineStepsArgs struct {
	Workspace    string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug     string `json:"repo_slug" jsonschema:"Repository slug"`
	PipelineUUID string `json:"pipeline_uuid" jsonschema:"Pipeline UUID"`
}

// ListPipelineStepsHandler lists steps in a pipeline.
func (c *Client) ListPipelineStepsHandler(ctx context.Context, req *mcp.CallToolRequest, args ListPipelineStepsArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.PipelineUUID == "" {
		return ToolResultError("workspace, repo_slug, and pipeline_uuid are required"), nil, nil
	}

	result, err := GetPaginated[PipelineStep](c, fmt.Sprintf("/repositories/%s/%s/pipelines/%s/steps",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.PipelineUUID))
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to list pipeline steps: %v", err)), nil, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ToolResultText(string(data)), nil, nil
}

type GetPipelineStepLogArgs struct {
	Workspace    string `json:"workspace" jsonschema:"Workspace slug"`
	RepoSlug     string `json:"repo_slug" jsonschema:"Repository slug"`
	PipelineUUID string `json:"pipeline_uuid" jsonschema:"Pipeline UUID"`
	StepUUID     string `json:"step_uuid" jsonschema:"Step UUID"`
}

// GetPipelineStepLogHandler gets the log output for a pipeline step.
func (c *Client) GetPipelineStepLogHandler(ctx context.Context, req *mcp.CallToolRequest, args GetPipelineStepLogArgs) (*mcp.CallToolResult, any, error) {
	if args.Workspace == "" || args.RepoSlug == "" || args.PipelineUUID == "" || args.StepUUID == "" {
		return ToolResultError("workspace, repo_slug, pipeline_uuid, and step_uuid are required"), nil, nil
	}

	raw, _, err := c.GetRaw(fmt.Sprintf("/repositories/%s/%s/pipelines/%s/steps/%s/log",
		QueryEscape(args.Workspace), QueryEscape(args.RepoSlug), args.PipelineUUID, args.StepUUID))
	if err != nil {
		return ToolResultError(fmt.Sprintf("failed to get step log: %v", err)), nil, nil
	}

	return ToolResultText(string(raw)), nil, nil
}
