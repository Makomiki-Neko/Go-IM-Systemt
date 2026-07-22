package friendservicelogic

import (
	"context"
	"errors"

	"IMM/common/models"
	"IMM/rpc/relation/internal/svc"
	"IMM/rpc/relation/relation"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeleteFriendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFriendLogic {
	return &DeleteFriendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteFriendLogic) DeleteFriend(in *relation.DeleteFriendReq) (*relation.CommonResponse, error) {
	// 软删除
	_, err := gorm.G[models.UserFriend](l.svcCtx.DB).Where("user_id = ? AND friend_id = ?", in.UserId, in.FriendId).Delete(l.ctx)

	// 在对方好友关系里标记删除
	_, err = gorm.G[models.UserFriend](l.svcCtx.DB).Where("user_id = ? AND friend_id = ?", in.FriendId, in.UserId).Updates(l.ctx, models.UserFriend{Status: 3})

	if err != nil {
		return nil, errors.New("DB Failed; " + err.Error())
	}

	return &relation.CommonResponse{}, nil
}
