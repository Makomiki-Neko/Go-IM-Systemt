// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"

	"IMM/api/internal/svc"
	"IMM/api/internal/types"
	"IMM/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateEmailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateEmailLogic {
	return &UpdateEmailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateEmailLogic) UpdateEmail(req *types.UpdateEmailRequest) (resp *types.UpdateEmailResponse, err error) {
	r, err := l.svcCtx.UserRpc.UpdateEmail(l.ctx, &user.UpdateEmailReq{
		Account: req.Account,
		Email:   req.Email,
	})
	if err != nil {
		return nil, err
	}
	return &types.UpdateEmailResponse{Code: int(r.Code)}, nil
}
