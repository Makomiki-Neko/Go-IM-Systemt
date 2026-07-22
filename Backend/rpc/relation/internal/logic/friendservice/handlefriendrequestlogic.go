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

type HandleFriendRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandleFriendRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleFriendRequestLogic {
	return &HandleFriendRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 处理好友请求
func (l *HandleFriendRequestLogic) HandleFriendRequest(in *relation.HandleFriendRequestReq) (*relation.HandleFriendRequestResp, error) {
	if in.Accept {
		//fmt.Printf("__R___:%v", in)
		_, err := gorm.G[models.UserFriend](l.svcCtx.DB).Select("status", "remark").Where("user_id = ? AND friend_id = ?", in.SenderId, in.HandlerId).Updates(l.ctx, models.UserFriend{Status: 1, Remark: ""}) // 使用Select强制更新字段
		if err != nil {
			return nil, errors.New("DB Updates Failed; " + err.Error())
		}
		// 双向记录
		err = gorm.G[models.UserFriend](l.svcCtx.DB).Create(l.ctx, &models.UserFriend{UserID: in.HandlerId, FriendID: in.SenderId, Status: 1})
		if err != nil {
			return nil, errors.New("DB Updates Failed; " + err.Error())
		}
		// 消息队列通知

		return &relation.HandleFriendRequestResp{Base: &relation.CommonResponse{Code: 200, Msg: "Friend Request Accepted."}}, nil
	}

	// 拒绝
	_, err := gorm.G[models.UserFriend](l.svcCtx.DB).Where("id = ?", in.RequestId).Updates(l.ctx, models.UserFriend{Status: 2})
	if err != nil {
		return nil, errors.New("DB Updates Failed; " + err.Error())
	}

	// 调用对话发送一句话

	return &relation.HandleFriendRequestResp{Base: &relation.CommonResponse{Code: 200, Msg: "Friend Request Refused."}}, nil
}
