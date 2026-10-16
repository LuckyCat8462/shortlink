package middleware

import (
	"net/http"

	"shortlink/internal/svc"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ApiKeyMiddleware(svcCtx *svc.ServiceContext) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key == "" || key != svcCtx.Config.ApiKey {
				httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized, map[string]any{
					"code": http.StatusUnauthorized,
					"msg":  "unauthorized",
				})
				return
			}
			next(w, r)
		}
	}
}

func CreateRateLimitMiddleware(svcCtx *svc.ServiceContext) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key == "" {
				key = "anonymous"
			}
			code, err := svcCtx.CreateLimiter.Take(key)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
				return
			}
			if code == limit.OverQuota {
				httpx.WriteJsonCtx(r.Context(), w, http.StatusTooManyRequests, map[string]any{
					"code": http.StatusTooManyRequests,
					"msg":  "rate limit exceeded",
				})
				return
			}
			next(w, r)
		}
	}
}

func RedirectRateLimitMiddleware(svcCtx *svc.ServiceContext) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := httpx.GetRemoteAddr(r)
			code, err := svcCtx.RedirectLimiter.Take(ip)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if code == limit.OverQuota {
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			next(w, r)
		}
	}
}

func AssistantRateLimitMiddleware(svcCtx *svc.ServiceContext) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key == "" {
				key = "anonymous"
			}
			code, err := svcCtx.AssistantLimiter.Take(key)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
				return
			}
			if code == limit.OverQuota {
				httpx.WriteJsonCtx(r.Context(), w, http.StatusTooManyRequests, map[string]any{
					"code": http.StatusTooManyRequests,
					"msg":  "rate limit exceeded",
				})
				return
			}
			next(w, r)
		}
	}
}
