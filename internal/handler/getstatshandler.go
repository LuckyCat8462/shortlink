package handler

import (
	"net/http"

	"shortlink/internal/logic"
	"shortlink/internal/svc"
	"shortlink/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetStatsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetStatsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, map[string]any{
				"code": http.StatusBadRequest,
				"msg":  err.Error(),
			})
			return
		}

		l := logic.NewGetStatsLogic(r.Context(), svcCtx)
		resp, err := l.GetStats(&req)
		if err != nil {
			writeAPIError(w, r, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
