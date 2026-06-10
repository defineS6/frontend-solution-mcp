package main

import (
	"context"
	"strings"
)

const frontendSystemPrompt = `你是资深前端方案顾问，擅长产品交互、信息架构、视觉风格、响应式布局、组件拆分和工程落地。请使用中文回答，优先给出可执行的前端方案、关键设计取舍、组件结构、交互状态和实现步骤。不要编写复杂后端业务逻辑。`

type FrontendService struct {
	defaultModel string
	client       ModelClient
	sessions     *SessionStore
}

type SolutionRequest struct {
	Prompt            string
	SessionID         string
	Model             string
	ReturnAllMessages bool
}

type ToolResponse struct {
	Success       bool          `json:"success"`
	SessionID     string        `json:"SESSION_ID,omitempty"`
	AgentMessages string        `json:"agent_messages,omitempty"`
	AllMessages   []ChatMessage `json:"all_messages,omitempty"`
	Error         string        `json:"error,omitempty"`
}

func NewFrontendService(defaultModel string, client ModelClient, sessions *SessionStore) *FrontendService {
	return &FrontendService{
		defaultModel: defaultModel,
		client:       client,
		sessions:     sessions,
	}
}

func (s *FrontendService) Generate(ctx context.Context, req SolutionRequest) ToolResponse {
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return failureResponse("PROMPT 不能为空")
	}

	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = s.defaultModel
	}
	if strings.TrimSpace(model) == "" {
		return failureResponse("模型名不能为空")
	}

	userMessage := ChatMessage{Role: "user", Content: prompt}
	sessionID, historyWithPrompt, err := s.sessions.MessagesForRequest(req.SessionID, userMessage)
	if err != nil {
		return failureResponse(err.Error())
	}

	modelMessages := append([]ChatMessage{{Role: "system", Content: frontendSystemPrompt}}, historyWithPrompt...)
	answer, err := s.client.Complete(ctx, model, modelMessages)
	if err != nil {
		return failureResponse(err.Error())
	}

	allMessages := s.sessions.AppendExchange(sessionID, userMessage, ChatMessage{Role: "assistant", Content: answer})
	resp := ToolResponse{
		Success:       true,
		SessionID:     sessionID,
		AgentMessages: answer,
	}
	if req.ReturnAllMessages {
		resp.AllMessages = allMessages
	}
	return resp
}

func failureResponse(message string) ToolResponse {
	return ToolResponse{
		Success: false,
		Error:   message,
	}
}
