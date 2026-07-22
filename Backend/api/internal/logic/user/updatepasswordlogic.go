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

type UpdatePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdatePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePasswordLogic {
	return &UpdatePasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatePasswordLogic) UpdatePassword(req *types.UpdatePasswordRequest) (resp *types.UpdatePasswordResponse, err error) {
	r, err := l.svcCtx.UserRpc.UpdatePassword(l.ctx, &user.UpdatePasswordReq{
		Account:  req.Account,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &types.UpdatePasswordResponse{Code: int(r.Code)}, nil
}
