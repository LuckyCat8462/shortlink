package assistant

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

// LLMConfig is DeepSeek / OpenAI-compatible chat model settings.
type LLMConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// LLMConfigFromEnv reads SHORTLINK_LLM_* environment variables.
func LLMConfigFromEnv() LLMConfig {
	base := strings.TrimSpace(os.Getenv("SHORTLINK_LLM_BASE_URL"))
	if base == "" {
		base = "https://api.deepseek.com"
	}
	return LLMConfig{
		APIKey:  strings.TrimSpace(os.Getenv("SHORTLINK_LLM_API_KEY")),
		BaseURL: base,
		Model:   strings.TrimSpace(os.Getenv("SHORTLINK_LLM_MODEL")),
	}
}

// NewChatModel returns a ToolCalling-capable chat model from env (CLI helpers).
func NewChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	return NewChatModelWithConfig(ctx, LLMConfigFromEnv())
}

// NewChatModelWithConfig builds the chat model from explicit config.
func NewChatModelWithConfig(ctx context.Context, cfg LLMConfig) (model.ToolCallingChatModel, error) {
	if cfg.APIKey == "" || cfg.Model == "" {
		return nil, fmt.Errorf("SHORTLINK_LLM_API_KEY / SHORTLINK_LLM_MODEL required")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.deepseek.com"
	}
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
	})
}

// Enabled reports whether LLM credentials are present.
func (c LLMConfig) Enabled() bool {
	return c.APIKey != "" && c.Model != ""
}
