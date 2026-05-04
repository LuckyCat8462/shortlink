// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"github.com/go-playground/validator/v10"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"shorturl/internal/logic"
	"shorturl/internal/svc"
	"shorturl/internal/types"
)

func ShorturlHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ConvertRequest
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

		l := logic.NewShorturlLogic(r.Context(), svcCtx)
		resp, err := l.Shorturl(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
