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

type GetFriendRequestListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFriendRequestListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendRequestListLogic {
	return &GetFriendRequestListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFriendRequestListLogic) GetFriendRequestList(req *types.GetFriendApplyReq) (resp *types.GetFriendApplyResp, err error) {
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

	r, err := l.svcCtx.FriendRPC.GetFriendRequests(l.ctx, &relation.GetFriendRequestsReq{
		UserId: uint64(uid),
		Page: &relation.PageRequest{
			Page: req.Page,
			Size: req.Size,
		},
	})
	if err != nil {
		return nil, err
	}
	returnList := make([]types.FriendInfo, 0, len(r.List))
	for _, v := range r.List {
		returnList = append(returnList, types.FriendInfo{
			Id:          v.Id,
			FriendName:  v.FriendName,
			Remark:      v.Remark,
			Avatar:      v.Avatar,
			Signature:   v.Signature,
			Gender:      v.Gender,
			Birthday:    v.Birthday,
			LastOnline:  v.LastOnline,
			OnlineStatu: v.OnlineStatu,
			Status:      int32(v.Status),
			CreatedAt:   v.CreatedAt,
		})
	}
	return &types.GetFriendApplyResp{
		CommonResp: types.CommonResp{Code: r.Base.Code, Msg: r.Base.Msg},
		PageResp:   types.PageResp{Page: r.Page.Page, Size: r.Page.Size, Total: r.Page.Total},
		List:       returnList,
	}, nil
}
