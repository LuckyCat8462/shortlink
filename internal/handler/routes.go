package handler

import (
	"net/http"

	"shortlink/internal/middleware"
	"shortlink/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	apiKey := middleware.ApiKeyMiddleware(serverCtx)
	createLimit := middleware.CreateRateLimitMiddleware(serverCtx)
	assistantLimit := middleware.AssistantRateLimitMiddleware(serverCtx)
	redirectLimit := middleware.RedirectRateLimitMiddleware(serverCtx)

	// Demo UI + health before parametric /:shortCode
	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/",
			Handler: serveWebFile("index.html"),
		},
		{
			Method:  http.MethodGet,
			Path:    "/static/app.css",
			Handler: serveWebFile("app.css"),
		},
		{
			Method:  http.MethodGet,
			Path:    "/static/app.js",
			Handler: serveWebFile("app.js"),
		},
		{
			Method: http.MethodGet,
			Path:   "/health",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			},
		},
	})

	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{apiKey, createLimit},
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/links",
				Handler: CreateLinkHandler(serverCtx),
			},
		),
		rest.WithPrefix("/api/v1"),
	)

	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{apiKey},
			rest.Route{
				Method:  http.MethodGet,
				Path:    "/links/:shortCode",
				Handler: GetLinkHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodGet,
				Path:    "/links/:shortCode/stats",
				Handler: GetStatsHandler(serverCtx),
			},
		),
		rest.WithPrefix("/api/v1"),
	)

	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{apiKey, assistantLimit},
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/assistant/chat",
				Handler: AssistantChatHandler(serverCtx),
			},
		),
		rest.WithPrefix("/api/v1"),
	)

	// Parametric redirect last
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{redirectLimit},
			rest.Route{
				Method:  http.MethodGet,
				Path:    "/:shortCode",
				Handler: RedirectHandler(serverCtx),
			},
		),
	)
}
