package main

import (
	"context"
	"errors"
	"testing"
)

func TestFrontendServiceGenerateUsesDefaultModel(t *testing.T) {
	client := &fakeModelClient{answer: "前端方案"}
	service := NewFrontendService("default-model", client, NewSessionStore())

	resp := service.Generate(context.Background(), SolutionRequest{
		Prompt: "设计一个天气卡片",
	})
	if !resp.Success {
		t.Fatalf("success = false, error = %q", resp.Error)
	}
	if resp.AgentMessages != "前端方案" {
		t.Fatalf("agent_messages = %q", resp.AgentMessages)
	}
	if resp.SessionID == "" {
		t.Fatal("SESSION_ID is empty")
	}
	if client.gotModel != "default-model" {
		t.Fatalf("model = %q", client.gotModel)
	}
	if len(client.gotMessages) != 2 {
		t.Fatalf("messages len = %d", len(client.gotMessages))
	}
	if client.gotMessages[0].Role != "system" {
		t.Fatalf("first role = %q", client.gotMessages[0].Role)
	}
	if client.gotMessages[1].Content != "设计一个天气卡片" {
		t.Fatalf("user prompt = %q", client.gotMessages[1].Content)
	}
}

func TestFrontendServiceGenerateModelOverrideAndAllMessages(t *testing.T) {
	client := &fakeModelClient{answer: "覆盖模型方案"}
	service := NewFrontendService("default-model", client, NewSessionStore())

	resp := service.Generate(context.Background(), SolutionRequest{
		Prompt:            "做一个后台首页",
		Model:             "override-model",
		ReturnAllMessages: true,
	})
	if !resp.Success {
		t.Fatalf("success = false, error = %q", resp.Error)
	}
	if client.gotModel != "override-model" {
		t.Fatalf("model = %q", client.gotModel)
	}
	if len(resp.AllMessages) != 2 {
		t.Fatalf("all_messages len = %d", len(resp.AllMessages))
	}
	if resp.AllMessages[0].Role != "user" || resp.AllMessages[1].Role != "assistant" {
		t.Fatalf("all_messages = %#v", resp.AllMessages)
	}
}

func TestFrontendServiceGenerateContinuesSession(t *testing.T) {
	client := &fakeModelClient{answer: "第一版"}
	service := NewFrontendService("default-model", client, NewSessionStore())

	first := service.Generate(context.Background(), SolutionRequest{Prompt: "先出方案"})
	if !first.Success {
		t.Fatalf("first success = false, error = %q", first.Error)
	}

	client.answer = "第二版"
	second := service.Generate(context.Background(), SolutionRequest{
		Prompt:    "继续优化",
		SessionID: first.SessionID,
	})
	if !second.Success {
		t.Fatalf("second success = false, error = %q", second.Error)
	}
	if second.SessionID != first.SessionID {
		t.Fatalf("session = %q, want %q", second.SessionID, first.SessionID)
	}
	if len(client.gotMessages) != 4 {
		t.Fatalf("messages len = %d, want 4", len(client.gotMessages))
	}
	if client.gotMessages[1].Content != "先出方案" || client.gotMessages[2].Content != "第一版" || client.gotMessages[3].Content != "继续优化" {
		t.Fatalf("messages = %#v", client.gotMessages)
	}
}

func TestFrontendServiceGenerateFailure(t *testing.T) {
	service := NewFrontendService("default-model", &fakeModelClient{err: errors.New("模型错误")}, NewSessionStore())

	resp := service.Generate(context.Background(), SolutionRequest{Prompt: "生成方案"})
	if resp.Success {
		t.Fatal("success = true")
	}
	if resp.Error != "模型错误" {
		t.Fatalf("error = %q", resp.Error)
	}
}

type fakeModelClient struct {
	answer      string
	err         error
	gotModel    string
	gotMessages []ChatMessage
}

func (f *fakeModelClient) Complete(ctx context.Context, model string, messages []ChatMessage) (string, error) {
	f.gotModel = model
	f.gotMessages = cloneMessages(messages)
	if f.err != nil {
		return "", f.err
	}
	return f.answer, nil
}
