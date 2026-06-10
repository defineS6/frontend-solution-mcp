package main

import (
	"strings"
	"testing"
	"time"
)

func TestLoadConfigFromEnv(t *testing.T) {
	cfg, err := LoadConfigFromEnv(mapLookup(map[string]string{
		envBaseURL: " https://api.example.com/v1 ",
		envAPIKey:  " key ",
		envModel:   " model-a ",
		envProxy:   " http://127.0.0.1:7890 ",
		envTimeout: "2m",
	}))
	if err != nil {
		t.Fatalf("LoadConfigFromEnv() error = %v", err)
	}

	if cfg.BaseURL != "https://api.example.com/v1" {
		t.Fatalf("BaseURL = %q", cfg.BaseURL)
	}
	if cfg.APIKey != "key" {
		t.Fatalf("APIKey = %q", cfg.APIKey)
	}
	if cfg.Model != "model-a" {
		t.Fatalf("Model = %q", cfg.Model)
	}
	if cfg.ProxyURL != "http://127.0.0.1:7890" {
		t.Fatalf("ProxyURL = %q", cfg.ProxyURL)
	}
	if cfg.Timeout != 2*time.Minute {
		t.Fatalf("Timeout = %s", cfg.Timeout)
	}
}

func TestLoadConfigDefaultTimeout(t *testing.T) {
	cfg, err := LoadConfigFromEnv(mapLookup(map[string]string{
		envBaseURL: "https://api.example.com/v1",
		envAPIKey:  "key",
		envModel:   "model-a",
	}))
	if err != nil {
		t.Fatalf("LoadConfigFromEnv() error = %v", err)
	}
	if cfg.Timeout != defaultTimeout {
		t.Fatalf("Timeout = %s, want %s", cfg.Timeout, defaultTimeout)
	}
}

func TestLoadConfigRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{
			name: "missing base url",
			env: map[string]string{
				envAPIKey: "key",
				envModel:  "model-a",
			},
			wantErr: envBaseURL,
		},
		{
			name: "missing api key",
			env: map[string]string{
				envBaseURL: "https://api.example.com/v1",
				envModel:   "model-a",
			},
			wantErr: envAPIKey,
		},
		{
			name: "missing model",
			env: map[string]string{
				envBaseURL: "https://api.example.com/v1",
				envAPIKey:  "key",
			},
			wantErr: envModel,
		},
		{
			name: "invalid timeout",
			env: map[string]string{
				envBaseURL: "https://api.example.com/v1",
				envAPIKey:  "key",
				envModel:   "model-a",
				envTimeout: "soon",
			},
			wantErr: envTimeout,
		},
		{
			name: "zero timeout",
			env: map[string]string{
				envBaseURL: "https://api.example.com/v1",
				envAPIKey:  "key",
				envModel:   "model-a",
				envTimeout: "0s",
			},
			wantErr: envTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadConfigFromEnv(mapLookup(tt.env))
			if err == nil {
				t.Fatal("LoadConfigFromEnv() error = nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want contains %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
