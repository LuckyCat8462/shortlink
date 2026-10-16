package assistant

import (
	"context"
	"fmt"
	"strings"
	"time"

	"shortlink/internal/stats"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

// NewStatsAgent builds a ReAct agent bound to stats tools + system prompt.
func NewStatsAgent(ctx context.Context, repo *stats.Repo, llm LLMConfig) (*react.Agent, error) {
	if repo == nil {
		return nil, fmt.Errorf("nil stats repo")
	}

	cm, err := NewChatModelWithConfig(ctx, llm)
	if err != nil {
		return nil, err
	}
	tools, err := NewStatsTools(repo)
	if err != nil {
		return nil, err
	}

	persona := SystemPrompt(time.Now())
	return react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: cm,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: tools,
		},
		MessageModifier: react.NewPersonaModifier(persona),
		MaxStep:         8,
	})
}

// Ask runs one user question with optional history through the stats agent.
func Ask(ctx context.Context, ag *react.Agent, history []*schema.Message, userMessage string) (*schema.Message, error) {
	if ag == nil {
		return nil, fmt.Errorf("nil agent")
	}
	msgs := make([]*schema.Message, 0, len(history)+1)
	msgs = append(msgs, history...)
	msgs = append(msgs, &schema.Message{Role: schema.User, Content: userMessage})
	return ag.Generate(ctx, msgs)
}

// AskStream runs the full tool-calling loop via Generate (reliable with DeepSeek),
// then emits the final answer as SSE-friendly text chunks.
//
// Note: Agent.Stream + default StreamToolCallChecker can stop early when the model
// narrates in plain text before emitting tool_calls (exactly the failure mode we hit).
func AskStream(ctx context.Context, ag *react.Agent, history []*schema.Message, userMessage string, onDelta func(string) error) (string, error) {
	resp, err := Ask(ctx, ag, history, userMessage)
	if err != nil {
		return "", err
	}
	text := strings.TrimSpace(resp.Content)
	if text == "" {
		return "", nil
	}
	if onDelta == nil {
		return text, nil
	}
	if err := emitChunks(text, onDelta); err != nil {
		return text, err
	}
	return text, nil
}

func emitChunks(text string, onDelta func(string) error) error {
	const maxRunes = 24
	var buf strings.Builder
	n := 0
	for _, r := range text {
		buf.WriteRune(r)
		n++
		if n >= maxRunes || r == '\n' {
			if err := onDelta(buf.String()); err != nil {
				return err
			}
			buf.Reset()
			n = 0
		}
	}
	if buf.Len() > 0 {
		return onDelta(buf.String())
	}
	return nil
}
