// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"encoding/json"
	"errors"

	"IMM/api/internal/svc"
	"IMM/api/internal/types"
	"IMM/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateInfoLogic {
	return &UpdateInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateInfoLogic) UpdateInfo(req *types.UpdateUserInfoRequest) (resp *types.UpdateUserInfoResponse, err error) {
	uidVal := l.ctx.Value("uid")
	if uidVal == nil {
		return nil, errors.New("uid not found in context")
	}

	uidNum, ok := uidVal.(json.Number)
	if !ok {
		return nil, errors.New("invalid uid type in context")
	}

	uid, err := uidNum.Int64()
	if err != nil {
		return nil, errors.New("invalid uid value")
	}

	r, err := l.svcCtx.UserRpc.UpdateUserInfo(l.ctx, &user.InfoUpdateReq{
		Uid:       uint64(uid),
		Name:      req.Name,
		Gender:    req.Gender,
		Photo:     req.Photo,
		Signature: req.Signature,
		Birthday:  req.Birthday,
	})
	if err != nil {
		return nil, err
	}
	return &types.UpdateUserInfoResponse{Code: int(r.Code)}, nil
}
