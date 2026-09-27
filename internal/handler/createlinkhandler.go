package handler

import (
	"net/http"

	"shortlink/internal/logic"
	"shortlink/internal/svc"
	"shortlink/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateLinkHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateLinkReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, map[string]any{
				"code": http.StatusBadRequest,
				"msg":  err.Error(),
			})
			return
		}

		l := logic.NewCreateLinkLogic(r.Context(), svcCtx)
		resp, err := l.CreateLink(&req)
		if err != nil {
			writeAPIError(w, r, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func writeAPIError(w http.ResponseWriter, r *http.Request, err error) {
	if code, ok := logic.HTTPStatusFromError(err); ok {
		httpx.WriteJsonCtx(r.Context(), w, code, map[string]any{
			"code": code,
			"msg":  err.Error(),
		})
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusInternalServerError, map[string]any{
		"code": http.StatusInternalServerError,
		"msg":  err.Error(),
	})
}
