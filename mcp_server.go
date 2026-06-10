package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const version = "0.1.0"

func NewFrontendMCPServer(service *FrontendService) *server.MCPServer {
	s := server.NewMCPServer(
		"frontend-mcp",
		version,
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	tool := mcp.NewTool("frontend_solution",
		mcp.WithDescription("调用外部大模型生成前端方案、设计建议、组件拆分和实现思路。"),
		mcp.WithString("PROMPT",
			mcp.Required(),
			mcp.Description("要发送给外部大模型的前端需求或问题。"),
		),
		mcp.WithString("SESSION_ID",
			mcp.Description("可选会话 ID；为空时创建新会话。"),
		),
		mcp.WithString("model",
			mcp.Description("可选模型名；不填时使用 FRONTEND_MCP_MODEL。"),
		),
		mcp.WithBoolean("return_all_messages",
			mcp.Description("是否返回当前会话的完整消息历史。"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		resp := service.Generate(ctx, SolutionRequest{
			Prompt:            request.GetString("PROMPT", ""),
			SessionID:         request.GetString("SESSION_ID", ""),
			Model:             request.GetString("model", ""),
			ReturnAllMessages: request.GetBool("return_all_messages", false),
		})

		fallback := resp.AgentMessages
		if !resp.Success {
			fallback = resp.Error
		}
		result := mcp.NewToolResultStructured(resp, fallback)
		result.IsError = !resp.Success
		return result, nil
	})

	return s
}
