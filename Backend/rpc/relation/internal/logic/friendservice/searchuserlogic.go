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

type SearchUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchUserLogic {
	return &SearchUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SearchUserLogic) SearchUser(in *relation.SearchUserReq) (*relation.SearchUserResp, error) {
	if in.Page.Page < 1 {
		in.Page.Page = 1
	}
	key, starRecord := fmt.Sprintf("%%%s%%", in.SearchName), (in.Page.Page-1)*in.Page.Size
	// 暂用模糊查询	只查询用户自定义名称
	users, err := gorm.G[models.UserInfo](l.svcCtx.DB).Where("name LIKE ?", key).Offset(int(starRecord)).Limit(int(in.Page.Size)).Find(l.ctx)
	if err != nil {
		return nil, errors.New("DB Failed; " + err.Error())
	}
	userInfoList := make([]*relation.FriendInfo, 0, len(users))
	for _, item := range users {
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
		userInfoList = append(userInfoList, &relation.FriendInfo{
			Id:          item.UserID,
			FriendName:  item.Name,
			Avatar:      item.Photo,
			Signature:   item.Signature,
			Gender:      item.Gender,
			Birthday:    b,
			LastOnline:  last,
			OnlineStatu: online,
		})

	}

	return &relation.SearchUserResp{Base: &relation.CommonResponse{Code: 200, Msg: "Search Success."}, List: userInfoList, Page: &relation.PageInfo{Page: in.Page.Page, Size: in.Page.Size}}, nil
}
