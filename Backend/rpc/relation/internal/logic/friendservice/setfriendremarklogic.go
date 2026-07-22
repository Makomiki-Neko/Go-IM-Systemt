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

type SetFriendRemarkLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetFriendRemarkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetFriendRemarkLogic {
	return &SetFriendRemarkLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetFriendRemarkLogic) SetFriendRemark(in *relation.SetFriendRemarkReq) (*relation.CommonResponse, error) {
	// 要求好友状态为已通过才允许设置别名
	_, err := gorm.G[models.UserFriend](l.svcCtx.DB).Where("user_id = ? AND friend_id = ? AND status = 1", in.UserId, in.FriendId).Updates(l.ctx, models.UserFriend{
		Remark: in.Remark,
	})
	if err != nil {
		return nil, errors.New("DB Updates Failed; " + err.Error())
	}

	return &relation.CommonResponse{Code: 200, Msg: "Set NickName Success."}, nil
}
