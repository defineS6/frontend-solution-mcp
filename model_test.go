package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestChatCompletionsEndpoint(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    string
	}{
		{
			name:    "root",
			baseURL: "https://api.example.com",
			want:    "https://api.example.com/v1/chat/completions",
		},
		{
			name:    "v1",
			baseURL: "https://api.example.com/v1",
			want:    "https://api.example.com/v1/chat/completions",
		},
		{
			name:    "custom compatible path",
			baseURL: "https://api.example.com/compatible-mode/v1",
			want:    "https://api.example.com/compatible-mode/v1/chat/completions",
		},
		{
			name:    "full endpoint",
			baseURL: "https://api.example.com/v1/chat/completions",
			want:    "https://api.example.com/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := chatCompletionsEndpoint(tt.baseURL)
			if err != nil {
				t.Fatalf("chatCompletionsEndpoint() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("endpoint = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewHTTPTransportProxy(t *testing.T) {
	transport, err := newHTTPTransport("")
	if err != nil {
		t.Fatalf("newHTTPTransport() error = %v", err)
	}
	if transport.Proxy != nil {
		t.Fatal("Proxy should be nil when proxy url is empty")
	}

	transport, err = newHTTPTransport("http://127.0.0.1:7890")
	if err != nil {
		t.Fatalf("newHTTPTransport() error = %v", err)
	}

	targetURL := mustParseURL(t, "https://api.example.com/v1/chat/completions")
	proxyURL, err := transport.Proxy(&http.Request{URL: targetURL})
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}
	if proxyURL.String() != "http://127.0.0.1:7890" {
		t.Fatalf("proxy = %q", proxyURL.String())
	}
}

func TestNewHTTPTransportInvalidProxy(t *testing.T) {
	_, err := newHTTPTransport("socks5://127.0.0.1:7890")
	if err == nil {
		t.Fatal("newHTTPTransport() error = nil")
	}
	if !strings.Contains(err.Error(), "仅支持 http 或 https") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestOpenAIClientCompleteSuccess(t *testing.T) {
	var captured openAIChatRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("Authorization = %q", r.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"方案内容"}}]}`))
	}))
	defer server.Close()

	client, err := NewOpenAIClient(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("NewOpenAIClient() error = %v", err)
	}

	answer, err := client.Complete(context.Background(), "model-a", []ChatMessage{{Role: "user", Content: "做一个卡片"}})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if answer != "方案内容" {
		t.Fatalf("answer = %q", answer)
	}
	if captured.Model != "model-a" {
		t.Fatalf("model = %q", captured.Model)
	}
	if len(captured.Messages) != 1 || captured.Messages[0].Content != "做一个卡片" {
		t.Fatalf("messages = %#v", captured.Messages)
	}
}

func TestOpenAIClientCompleteErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"bad key"}}`))
	}))
	defer server.Close()

	client, err := NewOpenAIClient(Config{
		BaseURL: server.URL + "/v1",
		APIKey:  "test-key",
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("NewOpenAIClient() error = %v", err)
	}

	_, err = client.Complete(context.Background(), "model-a", []ChatMessage{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("Complete() error = nil")
	}
	if !strings.Contains(err.Error(), "bad key") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestOpenAIClientCompleteEmptyContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"   "}}]}`))
	}))
	defer server.Close()

	client, err := NewOpenAIClient(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("NewOpenAIClient() error = %v", err)
	}

	_, err = client.Complete(context.Background(), "model-a", []ChatMessage{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("Complete() error = nil")
	}
	if !strings.Contains(err.Error(), "模型返回内容为空") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestOpenAIClientCompleteTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"慢响应"}}]}`))
	}))
	defer server.Close()

	client, err := NewOpenAIClient(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: 10 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewOpenAIClient() error = %v", err)
	}

	_, err = client.Complete(context.Background(), "model-a", []ChatMessage{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("Complete() error = nil")
	}
	if !strings.Contains(err.Error(), "模型请求超时") {
		t.Fatalf("error = %q", err.Error())
	}
}

func mustParseURL(t *testing.T, rawURL string) *url.URL {
	t.Helper()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	return parsed
}
