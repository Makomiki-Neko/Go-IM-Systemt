package logic

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"IMM/common/models"
	"IMM/rpc/chat/chat"
	"IMM/rpc/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateChatMsgLastReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateChatMsgLastReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatMsgLastReadLogic {
	return &UpdateChatMsgLastReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新用户最后已读信息
func (l *UpdateChatMsgLastReadLogic) UpdateChatMsgLastRead(in *chat.UpdateLastReadMsgRep) (*chat.CommonResponse, error) {

	// 确保LastMsgID < 当前传入的已读ID，避免网络延迟情况下误更新
	r, err := gorm.G[models.UserFriend](l.svcCtx.DB).Where("user_id = ? AND friend_id = ? AND last_msg_id < ?", in.UserId, in.TargetId, in.MsgId).Updates(l.ctx, models.UserFriend{
		LastMsgID: in.MsgId,
	})
	// 没有被更新
	if r == 0 {
		return &chat.CommonResponse{Code: 100}, nil
	}
	if err != nil {
		return nil, errors.New("DB Updates Failed, err: " + err.Error())
	}

	// 更新Redis未读计数，存在并发问题，若有新消息在查询DB后入库并完成Redis自增，那么后续的Redis操作可能并不反映真实未读情况
	// 查询新消息数
	count, err := gorm.G[models.PrivateMessage](l.svcCtx.DB).Where("to_user_id = ? AND from_user_id = ? AND msg_id > ?", in.UserId, in.TargetId, in.MsgId).Count(l.ctx, "msg_id")
	key := fmt.Sprintf("im:chat:unread:private:%d", in.UserId)
	field := fmt.Sprintf("%d", in.TargetId)
	if count == 0 {
		// 清空Redis内对应未读计数
		_, err = l.svcCtx.Redis.Hdel(key, field)
	} else {
		err = l.svcCtx.Redis.Hset(key, field, strconv.FormatInt(count, 10))
	}
	if err != nil {
		// 处理错误，仅记录日志
		logx.Errorf("Redis Failed to delete field %s from key %s: %v", field, key, err)
	}

	return &chat.CommonResponse{Code: 100}, nil
}
