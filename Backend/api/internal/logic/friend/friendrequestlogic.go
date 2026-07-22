// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package friend

import (
	"context"
	"encoding/json"
	"errors"

	"IMM/api/internal/svc"
	"IMM/api/internal/types"
	"IMM/rpc/relation/relation"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendRequestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendRequestLogic {
	return &FriendRequestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FriendRequestLogic) FriendRequest(req *types.SendFriendReq) (resp *types.SendFriendResp, err error) {
	// Get Uid From JWT
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

	r, err := l.svcCtx.FriendRPC.SendFriendRequest(l.ctx, &relation.SendFriendRequestReq{
		UserId:   uint64(uid),
		FriendId: req.FriendId,
		Msg:      req.Msg,
	})
	if err != nil {
		return nil, err
	}
	return &types.SendFriendResp{CommonResp: types.CommonResp{Code: r.Base.Code, Msg: r.Base.Msg}, RequestId: r.RequestId}, nil
}
