// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"shorturl/internal/logic"
	"shorturl/internal/svc"
	"shorturl/internal/types"

	"github.com/go-playground/validator/v10"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ShowUrlHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ShowUrlRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 参数规则校验
		if err := validator.New().StructCtx(r.Context(), &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			logx.Error("validator check failed", logx.LogField{Key: "error", Value: err.Error()})
			return
		}

		l := logic.NewShowUrlLogic(r.Context(), svcCtx)
		resp, err := l.ShowUrl(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		} else {
			// httpx.OkJsonCtx(r.Context(), w, resp)
			http.Redirect(w, r, resp.LongURL, http.StatusFound)
		}
	}
}
