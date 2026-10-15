package config

import (
	"os"
	"strings"
)

var envMap = loadDotEnv()

// loadDotEnv reads a simple key=value .env file into a map.
// Environment variables set in the process take precedence at lookup time.
func loadDotEnv() map[string]string {
	m := map[string]string{}
	data, err := os.ReadFile(".env")
	if err != nil {
		return m
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		val = strings.Trim(val, `"'`)
		m[key] = val
	}
	return m
}

func get(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if v := envMap[key]; v != "" {
		return v
	}
	return fallback
}

// DeepSeek (text LLM)
var (
	DeepSeekAPIKey  = get("DEEPSEEK_API_KEY", "")
	DeepSeekBaseURL = get("DEEPSEEK_BASE_URL", "https://api.deepseek.com/v1")
	DeepSeekModel   = get("DEEPSEEK_MODEL", "deepseek-flash")
)

// MiMo (multimodal LLM)
var (
	MiMoAPIKey  = get("MIMO_API_KEY", "")
	MiMoBaseURL = get("MIMO_BASE_URL", "https://api.xiaomimimo.com/v1")
	MiMoModel   = get("MIMO_MODEL", "mimo-v2.6-flash")
)

// Server
var (
	ServerPort = get("SERVER_PORT", "8080")
	ServerHost = get("SERVER_HOST", "0.0.0.0")
)
