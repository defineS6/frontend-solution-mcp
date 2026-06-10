package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置错误：%v\n", err)
		os.Exit(1)
	}

	client, err := NewOpenAIClient(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化模型客户端失败：%v\n", err)
		os.Exit(1)
	}

	service := NewFrontendService(cfg.Model, client, NewSessionStore())
	mcpServer := NewFrontendMCPServer(service)

	if err := server.ServeStdio(mcpServer); err != nil {
		fmt.Fprintf(os.Stderr, "MCP 服务错误：%v\n", err)
		os.Exit(1)
	}
}
