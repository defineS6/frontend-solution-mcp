# Frontend MCP

一个用 Go 编写的 MCP 服务，用于通过 OpenAI 兼容接口调用外部大模型，生成前端方案、设计建议、组件拆分和实现思路。

## 配置

服务只读取 MCP 配置传入的环境变量，不读取本地 `.env` 文件。

| 环境变量 | 必填 | 说明 |
| --- | --- | --- |
| `FRONTEND_MCP_BASE_URL` | 是 | 外部模型服务地址，例如 `https://api.example.com/v1` |
| `FRONTEND_MCP_API_KEY` | 是 | 外部模型服务 key |
| `FRONTEND_MCP_MODEL` | 是 | 默认模型名 |
| `FRONTEND_MCP_PROXY_URL` | 否 | HTTP/HTTPS 代理地址，例如 `http://127.0.0.1:7890`；不填则直连 |
| `FRONTEND_MCP_TIMEOUT` | 否 | 请求超时时间，默认 `120s`，例如 `60s`、`2m` |

## MCP 示例

本地构建后使用：

```json
{
  "mcpServers": {
    "frontend-mcp": {
      "command": "frontend-mcp",
      "env": {
        "FRONTEND_MCP_BASE_URL": "https://api.example.com/v1",
        "FRONTEND_MCP_API_KEY": "你的 key",
        "FRONTEND_MCP_MODEL": "你的模型名",
        "FRONTEND_MCP_PROXY_URL": "http://127.0.0.1:7890",
        "FRONTEND_MCP_TIMEOUT": "120s"
      },
      "type": "stdio"
    }
  }
}
```

也可以上传到 GitHub 后用 Go 直接运行：

```json
{
  "mcpServers": {
    "frontend-mcp": {
      "command": "go",
      "args": [
        "run",
        "github.com/defineS6/frontend-solution-mcp@latest"
      ],
      "env": {
        "FRONTEND_MCP_BASE_URL": "https://api.example.com/v1",
        "FRONTEND_MCP_API_KEY": "你的 key",
        "FRONTEND_MCP_MODEL": "你的模型名",
        "FRONTEND_MCP_TIMEOUT": "120s"
      },
      "startup_timeout_sec": 300,
      "type": "stdio"
    }
  }
}
```

> 注意：首次运行 `go run github.com/defineS6/frontend-solution-mcp@latest` 时，Go 会下载源码并编译，建议保留较长的 `startup_timeout_sec`。

## 工具

工具名：`frontend_solution`

入参：

- `PROMPT`：必填，要发送给外部大模型的前端需求或问题。
- `SESSION_ID`：可选，会话 ID；为空时创建新会话。
- `model`：可选，覆盖 `FRONTEND_MCP_MODEL`。
- `return_all_messages`：可选，是否返回当前会话完整消息历史。

成功返回：

```json
{
  "success": true,
  "SESSION_ID": "session-xxx",
  "agent_messages": "模型返回的前端方案"
}
```

失败返回：

```json
{
  "success": false,
  "error": "错误信息"
}
```

## 开发

```bash
go test ./...
go build -o frontend-mcp .
```
