package llm

import (
	"sync"

	"github.com/sashabaranov/go-openai"

	"offerhunter/internal/config"
)

var (
	deepseekOnce sync.Once
	deepseekCli  *openai.Client

	mimoOnce sync.Once
	mimoCli  *openai.Client
)

// DeepSeek returns the shared text LLM client.
func DeepSeek() *openai.Client {
	deepseekOnce.Do(func() {
		cfg := openai.DefaultConfig(config.DeepSeekAPIKey)
		cfg.BaseURL = config.DeepSeekBaseURL
		deepseekCli = openai.NewClientWithConfig(cfg)
	})
	return deepseekCli
}

// MiMo returns the shared multimodal LLM client.
func MiMo() *openai.Client {
	mimoOnce.Do(func() {
		cfg := openai.DefaultConfig(config.MiMoAPIKey)
		cfg.BaseURL = config.MiMoBaseURL
		mimoCli = openai.NewClientWithConfig(cfg)
	})
	return mimoCli
}
