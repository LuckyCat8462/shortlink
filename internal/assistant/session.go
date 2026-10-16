package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"shortlink/pkg/rediskey"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	defaultMaxHistory = 20 // messages (user+assistant), ~10 turns
	defaultTTLSec     = 7200
)

type storedMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// SessionStore persists multi-turn chat history in Redis.
type SessionStore struct {
	rds        *redis.Redis
	maxHistory int
	ttlSec     int
}

func NewSessionStore(rds *redis.Redis, maxHistory, ttlSec int) *SessionStore {
	if maxHistory <= 0 {
		maxHistory = defaultMaxHistory
	}
	if ttlSec <= 0 {
		ttlSec = defaultTTLSec
	}
	return &SessionStore{rds: rds, maxHistory: maxHistory, ttlSec: ttlSec}
}

func NewSessionID() string {
	return uuid.NewString()
}

func (s *SessionStore) Load(ctx context.Context, sessionID string) ([]*schema.Message, error) {
	if s == nil || s.rds == nil || sessionID == "" {
		return nil, nil
	}
	raw, err := s.rds.GetCtx(ctx, rediskey.AssistantSession(sessionID))
	if err != nil {
		return nil, err
	}
	if raw == "" {
		return nil, nil
	}
	var stored []storedMsg
	if err := json.Unmarshal([]byte(raw), &stored); err != nil {
		return nil, fmt.Errorf("decode session: %w", err)
	}
	out := make([]*schema.Message, 0, len(stored))
	for _, m := range stored {
		role := schema.RoleType(m.Role)
		switch role {
		case schema.User, schema.Assistant:
			out = append(out, &schema.Message{Role: role, Content: m.Content})
		}
	}
	return out, nil
}

func (s *SessionStore) Append(ctx context.Context, sessionID string, userText, assistantText string) error {
	if s == nil || s.rds == nil || sessionID == "" {
		return fmt.Errorf("invalid session store")
	}
	hist, err := s.Load(ctx, sessionID)
	if err != nil {
		return err
	}
	hist = append(hist,
		&schema.Message{Role: schema.User, Content: userText},
		&schema.Message{Role: schema.Assistant, Content: assistantText},
	)
	if len(hist) > s.maxHistory {
		hist = hist[len(hist)-s.maxHistory:]
	}
	stored := make([]storedMsg, 0, len(hist))
	for _, m := range hist {
		stored = append(stored, storedMsg{Role: string(m.Role), Content: m.Content})
	}
	b, err := json.Marshal(stored)
	if err != nil {
		return err
	}
	return s.rds.SetexCtx(ctx, rediskey.AssistantSession(sessionID), string(b), s.ttlSec)
}

func NormalizeSessionID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return NewSessionID()
	}
	return id
}
