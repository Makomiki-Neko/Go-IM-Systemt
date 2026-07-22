// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package friend

import (
	"context"

	"IMM/api/internal/svc"
	"IMM/api/internal/types"
	"IMM/rpc/relation/relation"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchFriendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchFriendLogic {
	return &SearchFriendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchFriendLogic) SearchFriend(req *types.SearchUserReq) (resp *types.SearchUserResp, err error) {
	r, err := l.svcCtx.FriendRPC.SearchUser(l.ctx, &relation.SearchUserReq{SearchName: req.SearchName, Page: &relation.PageRequest{Page: req.Page, Size: req.Size}})
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
	return &types.SearchUserResp{
		CommonResp: types.CommonResp{Code: r.Base.Code, Msg: r.Base.Msg},
		List:       returnList,
		PageResp:   types.PageResp{Total: r.Page.Total, Page: r.Page.Page, Size: r.Page.Size},
	}, nil
}
