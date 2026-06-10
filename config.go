package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	envBaseURL = "FRONTEND_MCP_BASE_URL"
	envAPIKey  = "FRONTEND_MCP_API_KEY"
	envModel   = "FRONTEND_MCP_MODEL"
	envProxy   = "FRONTEND_MCP_PROXY_URL"
	envTimeout = "FRONTEND_MCP_TIMEOUT"

	defaultTimeout = 120 * time.Second
)

type Config struct {
	BaseURL  string
	APIKey   string
	Model    string
	ProxyURL string
	Timeout  time.Duration
}

func LoadConfig() (Config, error) {
	return LoadConfigFromEnv(os.LookupEnv)
}

func LoadConfigFromEnv(lookup func(string) (string, bool)) (Config, error) {
	cfg := Config{
		BaseURL:  lookupTrimmed(lookup, envBaseURL),
		APIKey:   lookupTrimmed(lookup, envAPIKey),
		Model:    lookupTrimmed(lookup, envModel),
		ProxyURL: lookupTrimmed(lookup, envProxy),
		Timeout:  defaultTimeout,
	}

	if cfg.BaseURL == "" {
		return Config{}, fmt.Errorf("%s 不能为空", envBaseURL)
	}
	if err := validateHTTPURL(cfg.BaseURL, envBaseURL); err != nil {
		return Config{}, err
	}
	if cfg.APIKey == "" {
		return Config{}, fmt.Errorf("%s 不能为空", envAPIKey)
	}
	if cfg.Model == "" {
		return Config{}, fmt.Errorf("%s 不能为空", envModel)
	}

	timeoutRaw := lookupTrimmed(lookup, envTimeout)
	if timeoutRaw != "" {
		timeout, err := time.ParseDuration(timeoutRaw)
		if err != nil {
			return Config{}, fmt.Errorf("%s 格式错误，请使用 Go duration 格式，例如 60s 或 2m：%w", envTimeout, err)
		}
		if timeout <= 0 {
			return Config{}, fmt.Errorf("%s 必须大于 0", envTimeout)
		}
		cfg.Timeout = timeout
	}

	return cfg, nil
}

func lookupTrimmed(lookup func(string) (string, bool), key string) string {
	value, ok := lookup(key)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func validateHTTPURL(rawURL string, field string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("%s 地址格式错误：%w", field, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("%s 仅支持 http 或 https 地址", field)
	}
	if parsed.Host == "" {
		return fmt.Errorf("%s 必须包含主机地址", field)
	}
	return nil
}
