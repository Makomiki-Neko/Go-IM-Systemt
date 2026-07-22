package friendservicelogic

import (
	"context"
	"errors"
	"fmt"

	"IMM/common/models"
	"IMM/rpc/relation/internal/svc"
	"IMM/rpc/relation/relation"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type GetFriendRequestsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFriendRequestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendRequestsLogic {
	return &GetFriendRequestsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取好友请求
func (l *GetFriendRequestsLogic) GetFriendRequests(in *relation.GetFriendRequestsReq) (*relation.GetFriendRequestsResp, error) {
	if in.Page.Page < 1 {
		in.Page.Page = 1
	} // 从第一页开始
	starRecord := (in.Page.Page - 1) * in.Page.Size
	friendReqList, err := gorm.G[models.UserFriend](l.svcCtx.DB).Select("user_id", "remark").Where("friend_id = ? AND status = 0", in.UserId).Offset(int(starRecord)).Limit(int(in.Page.Size)).Find(l.ctx)
	if err != nil {
		return nil, errors.New("DB Failed; " + err.Error())
	}

	// 提取所有请求方 user_id 组成 uint64 切片
	userIDs, idAndRemark := make([]uint64, 0, len(friendReqList)), map[uint64]string{}
	for _, item := range friendReqList {
		userIDs = append(userIDs, item.UserID)
		idAndRemark[item.UserID] = item.Remark // 请求方的留言
	}

	if len(userIDs) == 0 {
		return &relation.GetFriendRequestsResp{Base: &relation.CommonResponse{Code: 200, Msg: "No More Friend Req"}, List: nil, Page: &relation.PageInfo{Page: in.Page.Page, Size: in.Page.Size}}, nil
	}

	// 获取发送方信息
	senderInfo, err := gorm.G[models.UserInfo](l.svcCtx.DB).Where("user_id IN ?", userIDs).Find(l.ctx)
	if err != nil {
		return nil, errors.New("DB Failed; " + err.Error())
	}

	senderInfoList := make([]*relation.FriendInfo, 0, len(userIDs))
	for _, item := range senderInfo {
		r, e := l.svcCtx.Redis.Get(fmt.Sprintf("user:status:%d", item.UserID))
		online := false
		if r != "" && !errors.Is(e, redis.Nil) {
			online = true
		}
		var b, last int64
		if item.Birthday != nil {
			b = item.Birthday.Unix()
		}
		if item.LastOnline != nil {
			last = item.LastOnline.Unix()
		}
		senderInfoList = append(senderInfoList, &relation.FriendInfo{
			Id:          item.UserID,
			FriendName:  item.Name,
			Remark:      idAndRemark[item.UserID],
			Avatar:      item.Photo,
			Signature:   item.Signature,
			Gender:      item.Gender,
			Birthday:    b,
			LastOnline:  last,
			OnlineStatu: online,
		})

	}

	return &relation.GetFriendRequestsResp{Base: &relation.CommonResponse{Code: 200, Msg: "Get Friend Req Info Success."}, List: senderInfoList, Page: &relation.PageInfo{Page: in.Page.Page, Size: in.Page.Size}}, nil
}
