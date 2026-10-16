package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"shortlink/internal/assistant"
	"shortlink/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type assistantChatReq struct {
	SessionID string `json:"sessionId"`
	Message   string `json:"message"`
}

// AssistantChatHandler streams assistant replies via SSE.
func AssistantChatHandler(serverCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if serverCtx.Chat == nil || serverCtx.Config.LLM.APIKey == "" || serverCtx.Config.LLM.Model == "" {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"code":    http.StatusServiceUnavailable,
				"message": "assistant unavailable: configure SHORTLINK_LLM_API_KEY and SHORTLINK_LLM_MODEL",
			})
			return
		}

		var req assistantChatReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid json body"})
			return
		}
		if req.Message == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "message is required"})
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "streaming unsupported"})
			return
		}

		timeout := time.Duration(serverCtx.Config.AssistantTimeoutSec) * time.Second
		if timeout <= 0 {
			timeout = 90 * time.Second
		}
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		sid := assistant.NormalizeSessionID(req.SessionID)
		_ = writeSSE(w, flusher, "session", map[string]string{"sessionId": sid})
		_ = writeSSE(w, flusher, "status", map[string]string{"state": "querying"})

		result, err := serverCtx.Chat.RunStream(ctx, assistant.ChatRequest{
			SessionID: sid,
			Message:   req.Message,
		}, func(delta string) error {
			return writeSSE(w, flusher, "delta", map[string]string{"text": delta})
		})
		if err != nil {
			logx.Errorw("assistant chat failed", logx.LogField{Key: "err", Value: err.Error()})
			_ = writeSSE(w, flusher, "error", map[string]string{"message": err.Error()})
			_ = writeSSE(w, flusher, "done", map[string]string{"ok": "false"})
			return
		}
		_ = writeSSE(w, flusher, "done", map[string]any{"ok": true, "sessionId": result.SessionID})
	}
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, event string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, b); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
