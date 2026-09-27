package handler

import (
	"net/http"

	"shortlink/internal/logic"
	"shortlink/internal/svc"
	"shortlink/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetLinkHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetLinkReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, map[string]any{
				"code": http.StatusBadRequest,
				"msg":  err.Error(),
			})
			return
		}

		l := logic.NewGetLinkLogic(r.Context(), svcCtx)
		resp, err := l.GetLink(&req)
		if err != nil {
			writeAPIError(w, r, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
