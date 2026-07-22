package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"IMM/common/models"
	"IMM/common/types"
	"IMM/rpc/chat/chat"
	"IMM/rpc/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type SendGroupMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendGroupMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendGroupMessageLogic {
	return &SendGroupMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发送群聊消息
func (l *SendGroupMessageLogic) SendGroupMessage(in *chat.SendGroupMessageReq) (*chat.SendMessageResp, error) {
	// 查询Redis最近是否有处理过该消息
	// 1. 构建幂等Key
	idempotentKey := fmt.Sprintf("chat:msg:%d:%d", in.FromUserId, in.ClientMsgId)

	// 2. 原子操作：尝试设置 Key, Setnx 返回 true 表示设置成功（首次请求），返回 false 表示 Key 已存在（重复请求）
	success, err := l.svcCtx.Redis.Setnx(idempotentKey, "1")
	if err != nil {
		return nil, errors.New("Redis Error: " + err.Error())
	}

	// 3. 如果设置失败，说明是重复请求，直接返回成功（幂等返回）
	if !success {
		logx.Infof("duplicate request, clientMsgId: %v", in.ClientMsgId)
		return &chat.SendMessageResp{CommonResponse: &chat.CommonResponse{Code: 206, Data: nil}}, nil
	}
	l.svcCtx.Redis.Expire(idempotentKey, 1800) // 30分钟后过期

	// 鉴权
	_, err = gorm.G[models.GroupMember](l.svcCtx.DB).Where("group_id = ? AND user_id = ? AND status = 1", in.GroupId, in.FromUserId).First(l.ctx)
	if err != nil {
		return nil, errors.New("Not in Group, err: " + err.Error())
	}

	// 写入数据库
	msg := models.GroupMessage{
		FromUserID: in.FromUserId,
		GroupID:    in.GroupId,
		MsgType:    int8(in.MsgType),
		Content:    in.Content,
		SendTime:   time.Now(),
	}
	err = gorm.G[models.GroupMessage](l.svcCtx.DB).Create(l.ctx, &msg)
	if err != nil {
		return nil, errors.New("DB Error: " + err.Error())
	}

	// Redis记录用户全局未读消息数，WS收到客户端已读回应后清零（客户端内打开包含最新消息的对话即视为已读）
	/* 需要所有群内成员未读计数+1
	key := fmt.Sprintf("im:chat:unread:group:%d", in.)
	field := fmt.Sprintf("%d", in.GroupId)
	_, err = l.svcCtx.Redis.Hincrby(key, field, 1) // Hash存储 在当前用户下 有多少各用户发来的未读消息数
	if err != nil {
		log.Fatalf("Chat Redis Incr Unread Number Failed, err: %v", err.Error())
		err = nil
	}
	*/

	// 加入消息队列
	msgBytes, _ := json.Marshal(msg)
	pubMsg := types.MqMsg{
		Uid:  in.GroupId, // 发送给
		Type: "group_message",
		Data: msgBytes,
	}
	body, _ := json.Marshal(pubMsg) // json序列化

	err = l.svcCtx.RabbitMQ.Publish(l.ctx, "im.gateway.push.chat.group", body)
	if err != nil {
		return nil, errors.New("Push Failed, Error: " + err.Error())
	}

	return &chat.SendMessageResp{}, nil
}
