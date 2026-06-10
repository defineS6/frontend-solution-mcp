package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const maxErrorBodyBytes = 1 << 20

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ModelClient interface {
	Complete(ctx context.Context, model string, messages []ChatMessage) (string, error)
}

type OpenAIClient struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
}

func NewOpenAIClient(cfg Config) (*OpenAIClient, error) {
	transport, err := newHTTPTransport(cfg.ProxyURL)
	if err != nil {
		return nil, err
	}

	endpoint, err := chatCompletionsEndpoint(cfg.BaseURL)
	if err != nil {
		return nil, err
	}

	return &OpenAIClient{
		apiKey:   cfg.APIKey,
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
	}, nil
}

func newHTTPTransport(proxyRawURL string) (*http.Transport, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// 默认直连，避免无意读取 HTTP_PROXY/HTTPS_PROXY 等通用代理变量。
	transport.Proxy = nil

	proxyRawURL = strings.TrimSpace(proxyRawURL)
	if proxyRawURL == "" {
		return transport, nil
	}

	parsed, err := url.Parse(proxyRawURL)
	if err != nil {
		return nil, fmt.Errorf("%s 地址格式错误：%w", envProxy, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("%s 仅支持 http 或 https 代理地址", envProxy)
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("%s 必须包含代理主机地址", envProxy)
	}

	transport.Proxy = http.ProxyURL(parsed)
	return transport, nil
}

func chatCompletionsEndpoint(baseRawURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(baseRawURL))
	if err != nil {
		return "", fmt.Errorf("%s 地址格式错误：%w", envBaseURL, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("%s 仅支持 http 或 https 地址", envBaseURL)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("%s 必须包含主机地址", envBaseURL)
	}

	path := strings.TrimRight(parsed.Path, "/")
	switch {
	case strings.HasSuffix(path, "/chat/completions"):
		parsed.Path = path
	case path == "":
		parsed.Path = "/v1/chat/completions"
	default:
		parsed.Path = path + "/chat/completions"
	}

	return parsed.String(), nil
}

func (c *OpenAIClient) Complete(ctx context.Context, model string, messages []ChatMessage) (string, error) {
	body := openAIChatRequest{
		Model:    model,
		Messages: messages,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("构造模型请求失败：%w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("创建模型请求失败：%w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) || os.IsTimeout(err) {
			return "", fmt.Errorf("模型请求超时")
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			return "", fmt.Errorf("模型请求已取消")
		}
		return "", fmt.Errorf("模型请求失败：%w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := readOpenAIError(resp.Body)
		if message == "" {
			message = resp.Status
		}
		return "", fmt.Errorf("模型接口返回 %d：%s", resp.StatusCode, message)
	}

	var decoded openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf("解析模型响应失败：%w", err)
	}
	if len(decoded.Choices) == 0 {
		return "", fmt.Errorf("模型响应缺少 choices")
	}

	content := strings.TrimSpace(decoded.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("模型返回内容为空")
	}
	return content, nil
}

func readOpenAIError(body io.Reader) string {
	raw, err := io.ReadAll(io.LimitReader(body, maxErrorBodyBytes))
	if err != nil {
		return ""
	}

	var decoded struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &decoded); err == nil && strings.TrimSpace(decoded.Error.Message) != "" {
		return strings.TrimSpace(decoded.Error.Message)
	}
	return strings.TrimSpace(string(raw))
}

type openAIChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}
