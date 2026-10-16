package assistant

import (
	"context"
	"fmt"
	"strings"

	"shortlink/internal/stats"
)

// ChatService wires agent + session for HTTP/SSE.
type ChatService struct {
	Repo     *stats.Repo
	Sessions *SessionStore
	LLM      LLMConfig
}

type ChatRequest struct {
	SessionID string
	Message   string
}

type ChatResult struct {
	SessionID string
	Answer    string
}

// RunStream loads session, streams the answer, then persists the turn.
func (s *ChatService) RunStream(ctx context.Context, req ChatRequest, onDelta func(string) error) (*ChatResult, error) {
	if s == nil || s.Repo == nil {
		return nil, fmt.Errorf("chat service not configured")
	}
	if !s.LLM.Enabled() {
		return nil, fmt.Errorf("LLM not configured (SHORTLINK_LLM_API_KEY / SHORTLINK_LLM_MODEL)")
	}
	msg := strings.TrimSpace(req.Message)
	if msg == "" {
		return nil, fmt.Errorf("message is required")
	}
	sid := NormalizeSessionID(req.SessionID)

	hist, err := s.Sessions.Load(ctx, sid)
	if err != nil {
		return nil, fmt.Errorf("load session: %w", err)
	}

	ag, err := NewStatsAgent(ctx, s.Repo, s.LLM)
	if err != nil {
		return nil, err
	}

	answer, err := AskStream(ctx, ag, hist, msg, onDelta)
	if err != nil {
		return nil, err
	}
	if answer == "" {
		answer = "暂无数据"
		if onDelta != nil {
			_ = onDelta(answer)
		}
	}

	if s.Sessions != nil {
		if err := s.Sessions.Append(ctx, sid, msg, answer); err != nil {
			return nil, fmt.Errorf("save session: %w", err)
		}
	}
	return &ChatResult{SessionID: sid, Answer: answer}, nil
}
