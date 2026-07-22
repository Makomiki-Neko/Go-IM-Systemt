// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package friend

import (
	"net/http"

	"IMM/api/internal/logic/friend"
	"IMM/api/internal/svc"
	"IMM/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func SetFriendMarkHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SetRemarkReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := friend.NewSetFriendMarkLogic(r.Context(), svcCtx)
		resp, err := l.SetFriendMark(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
