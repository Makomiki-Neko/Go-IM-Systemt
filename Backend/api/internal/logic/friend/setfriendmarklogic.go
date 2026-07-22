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

type SetFriendMarkLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetFriendMarkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetFriendMarkLogic {
	return &SetFriendMarkLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetFriendMarkLogic) SetFriendMark(req *types.SetRemarkReq) (resp *types.CommonResp, err error) {
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

	r, err := l.svcCtx.FriendRPC.SetFriendRemark(l.ctx, &relation.SetFriendRemarkReq{
		UserId:   uint64(uid),
		FriendId: req.FriendId,
		Remark:   req.Remark,
	})
	if err != nil {
		return nil, err
	}

	return &types.CommonResp{Code: r.Code, Msg: r.Msg}, nil
}
