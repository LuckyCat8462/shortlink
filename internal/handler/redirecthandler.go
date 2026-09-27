package handler

import (
	"errors"
	"net/http"

	"shortlink/internal/logic"
	"shortlink/internal/svc"
	"shortlink/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func RedirectHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RedirectReq
		if err := httpx.Parse(r, &req); err != nil {
			http.NotFound(w, r)
			return
		}

		l := logic.NewRedirectLogic(r.Context(), svcCtx)
		longURL, err := l.Resolve(&req)
		if err != nil {
			if errors.Is(err, logic.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			http.NotFound(w, r)
			return
		}

		l.RecordClick(req.ShortCode, httpx.GetRemoteAddr(r))
		http.Redirect(w, r, longURL, http.StatusFound)
	}
}
