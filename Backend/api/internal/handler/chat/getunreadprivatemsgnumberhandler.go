// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"net/http"

	"IMM/api/internal/logic/chat"
	"IMM/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetUnreadPrivateMsgNumberHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := chat.NewGetUnreadPrivateMsgNumberLogic(r.Context(), svcCtx)
		resp, err := l.GetUnreadPrivateMsgNumber()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
