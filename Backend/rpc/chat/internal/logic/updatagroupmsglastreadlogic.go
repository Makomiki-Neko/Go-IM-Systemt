package logic

import (
	"context"
	"errors"
	"fmt"

	"IMM/common/models"
	"IMM/rpc/chat/chat"
	"IMM/rpc/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdataGroupMsgLastReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdataGroupMsgLastReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdataGroupMsgLastReadLogic {
	return &UpdataGroupMsgLastReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新群组最新消息
func (l *UpdataGroupMsgLastReadLogic) UpdataGroupMsgLastRead(in *chat.UpdateLastReadMsgRep) (*chat.CommonResponse, error) {
	// 清空Redis内对应未读计数
	key := fmt.Sprintf("im:chat:unread:group:%d", in.UserId)
	field := fmt.Sprintf("%d", in.TargetId)
	_, err := l.svcCtx.Redis.Hdel(key, field)
	if err != nil {
		// 处理错误，例如记录日志
		logx.Errorf("Redis Failed to delete field %s from key %s: %v", field, key, err)
		err = nil
	}

	_, err = gorm.G[models.GroupMember](l.svcCtx.DB).Where("user_id = ? AND group_id = ?", in.UserId, in.TargetId).Updates(l.ctx, models.GroupMember{
		LastMsgID: in.MsgId,
	})
	if err != nil {
		return nil, errors.New("DB Updates Failed, err: " + err.Error())
	}

	return &chat.CommonResponse{Code: 100}, nil
}
