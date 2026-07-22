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

type GetFriendListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFriendListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendListLogic {
	return &GetFriendListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取好友列表
func (l *GetFriendListLogic) GetFriendList(in *relation.GetFriendListReq) (*relation.GetFriendListResp, error) {
	if in.Page.Page < 1 {
		in.Page.Page = 1
	} // 从第一页开始
	starRecord := (in.Page.Page - 1) * in.Page.Size
	friendList, err := gorm.G[models.UserFriend](l.svcCtx.DB).Select("friend_id", "remark").Where("user_id = ? AND status = 1", in.UserId).Offset(int(starRecord)).Limit(int(in.Page.Size)).Find(l.ctx)
	if err != nil {
		return nil, errors.New("DB Failed; " + err.Error())
	}

	// 提取所有 friend_id 组成 uint64 切片
	friendIDs, idAndRemark := make([]uint64, 0, len(friendList)), map[uint64]string{}
	for _, item := range friendList {
		friendIDs = append(friendIDs, item.FriendID)
		idAndRemark[item.FriendID] = item.Remark
	}

	if len(friendIDs) == 0 {
		return &relation.GetFriendListResp{Base: &relation.CommonResponse{Code: 200, Msg: "No More Friend"}, List: nil, Page: &relation.PageInfo{Page: in.Page.Page, Size: in.Page.Size}}, nil
	}

	// 获取好友信息
	friendInfo, err := gorm.G[models.UserInfo](l.svcCtx.DB).Where("user_id IN ?", friendIDs).Find(l.ctx)
	if err != nil {
		return nil, errors.New("DB Failed; " + err.Error())
	}

	friendInfoList := make([]*relation.FriendInfo, 0, len(friendIDs))
	for _, item := range friendInfo {
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
		friendInfoList = append(friendInfoList, &relation.FriendInfo{
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

	return &relation.GetFriendListResp{
		Base: &relation.CommonResponse{Code: 200, Msg: "Get Friend Info Success."},
		List: friendInfoList,
		Page: &relation.PageInfo{Page: in.Page.Page, Size: in.Page.Size},
	}, nil
}
