package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// ListPipelinesHandler lists pipeline runs for a repository.
func (c *Client) ListPipelinesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	pagelen := req.GetInt("pagelen", 25)
	page := req.GetInt("page", 1)
	sort := req.GetString("sort", "-created_on")

	path := fmt.Sprintf("/repositories/%s/%s/pipelines?pagelen=%d&page=%d&sort=%s",
		QueryEscape(workspace), QueryEscape(repoSlug), pagelen, page, sort)

	status := req.GetString("status", "")
	if status != "" {
		path += "&status=" + status
	}

	result, err := GetPaginated[Pipeline](c, path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list pipelines: %v", err)), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// GetPipelineHandler gets details for a single pipeline run.
func (c *Client) GetPipelineHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	pipelineUUID, err := req.RequireString("pipeline_uuid")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	pipe, err := GetJSON[Pipeline](c, fmt.Sprintf("/repositories/%s/%s/pipelines/%s",
		QueryEscape(workspace), QueryEscape(repoSlug), pipelineUUID))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get pipeline: %v", err)), nil
	}

	data, _ := json.MarshalIndent(pipe, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// TriggerPipelineHandler triggers a new pipeline run.
func (c *Client) TriggerPipelineHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	refName, err := req.RequireString("ref_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	refType := req.GetString("ref_type", "branch")
	pattern := req.GetString("pattern", "")

	body := TriggerPipelineRequest{
		Target: PipeTriggerTarget{
			Type:    "pipeline_ref_target",
			RefType: refType,
			RefName: refName,
		},
	}

	if pattern != "" {
		body.Target.Selector = &PipelineSelector{
			Type:    "custom",
			Pattern: pattern,
		}
	}

	respData, err := c.Post(fmt.Sprintf("/repositories/%s/%s/pipelines",
		QueryEscape(workspace), QueryEscape(repoSlug)), body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to trigger pipeline: %v", err)), nil
	}

	var pipe Pipeline
	if err := json.Unmarshal(respData, &pipe); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse response: %v", err)), nil
	}

	data, _ := json.MarshalIndent(pipe, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// StopPipelineHandler stops a running pipeline.
func (c *Client) StopPipelineHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	pipelineUUID, err := req.RequireString("pipeline_uuid")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	_, err = c.Post(fmt.Sprintf("/repositories/%s/%s/pipelines/%s/stopPipeline",
		QueryEscape(workspace), QueryEscape(repoSlug), pipelineUUID), nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to stop pipeline: %v", err)), nil
	}

	return mcp.NewToolResultText("Pipeline stopped successfully"), nil
}

// ListPipelineStepsHandler lists steps in a pipeline.
func (c *Client) ListPipelineStepsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	pipelineUUID, err := req.RequireString("pipeline_uuid")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result, err := GetPaginated[PipelineStep](c, fmt.Sprintf("/repositories/%s/%s/pipelines/%s/steps",
		QueryEscape(workspace), QueryEscape(repoSlug), pipelineUUID))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list pipeline steps: %v", err)), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// GetPipelineStepLogHandler gets the log output for a pipeline step.
func (c *Client) GetPipelineStepLogHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := req.RequireString("workspace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repoSlug, err := req.RequireString("repo_slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	pipelineUUID, err := req.RequireString("pipeline_uuid")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	stepUUID, err := req.RequireString("step_uuid")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	raw, _, err := c.GetRaw(fmt.Sprintf("/repositories/%s/%s/pipelines/%s/steps/%s/log",
		QueryEscape(workspace), QueryEscape(repoSlug), pipelineUUID, stepUUID))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get step log: %v", err)), nil
	}

	return mcp.NewToolResultText(string(raw)), nil
}
