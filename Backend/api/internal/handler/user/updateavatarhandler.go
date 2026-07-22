// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"net/http"

	"IMM/api/internal/logic/user"
	"IMM/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func UpdateAvatarHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := user.NewUpdateAvatarLogic(r.Context(), svcCtx, r)
		resp, err := l.UpdateAvatar()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
