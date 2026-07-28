package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"IMM/common/models"
	"IMM/rpc/chat/chat"
	"IMM/rpc/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type SendPrivateMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendPrivateMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendPrivateMessageLogic {
	return &SendPrivateMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发送私聊消息
func (l *SendPrivateMessageLogic) SendPrivateMessage(in *chat.SendPrivateMessageReq) (*chat.SendMessageResp, error) {

	logx.Info("\n——————1：RPC：GetNetChatMsg：", in.Content)

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
	l.svcCtx.Redis.Expire(idempotentKey, 300) // 过期时间

	// 鉴权
	_, err = gorm.G[models.UserFriend](l.svcCtx.DB).Where("user_id = ? AND friend_id = ? AND status = 1", in.FromUserId, in.ToUserId).First(l.ctx)
	if err != nil {
		return nil, errors.New("Target Not you Friend, err: " + err.Error())
	}

	// 不同类型消息处理，文件类型需验证文件已在SeaweedFS上存在
	if in.MsgType != 1 {

	}

	// 写入数据库
	msg := models.PrivateMessage{
		FromUserID: in.FromUserId,
		ToUserID:   in.ToUserId,
		MsgType:    int8(in.MsgType),
		Content:    in.Content,
		SendTime:   time.Now(),
		Status:     0,
	}
	err = gorm.G[models.PrivateMessage](l.svcCtx.DB).Create(l.ctx, &msg)
	if err != nil {
		return nil, errors.New("DB Error: " + err.Error())
	}

	// Redis记录用户全局未读消息数，WS收到客户端已读回应后清零（客户端内打开包含最新消息的对话即视为已读）
	// 增加未读计数
	key := fmt.Sprintf("im:chat:unread:private:%d", in.ToUserId)
	field := fmt.Sprintf("%d", in.FromUserId)
	_, err = l.svcCtx.Redis.Hincrby(key, field, 1) // Hash存储 在当前用户下 有多少各用户发来的未读消息数
	if err != nil {
		logx.Errorf("Chat Redis Incr Unread Number Failed, err: %v", err.Error())
		err = nil
	}

	msgBytes, _ := json.Marshal(msg)

	return &chat.SendMessageResp{MsgId: msg.MsgID, SendTime: msg.SendTime.Unix(), CommonResponse: &chat.CommonResponse{Code: 200, Data: msgBytes}}, nil
}
