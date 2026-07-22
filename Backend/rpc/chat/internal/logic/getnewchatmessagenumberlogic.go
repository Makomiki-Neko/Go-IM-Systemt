package logic

import (
	"context"
	"fmt"
	"strconv"

	"IMM/rpc/chat/chat"
	"IMM/rpc/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNewChatMessageNumberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetNewChatMessageNumberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNewChatMessageNumberLogic {
	return &GetNewChatMessageNumberLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取新消息数量
func (l *GetNewChatMessageNumberLogic) GetNewChatMessageNumber(in *chat.GetChatNewMessagesNumberReq) (*chat.GetMessageNumberResp, error) {
	// 一次性获取所有发送方及其未读数量
	key := fmt.Sprintf("im:chat:unread:private:%d", in.UserId)
	result, err := l.svcCtx.Redis.Hgetall(key)
	if err != nil {
		return nil, err
	}
	returnList, unReadTotal := make([]*chat.Unread, 0, len(result)), 0
	for k, v := range result {
		unReadTotal++
		id, _ := strconv.Atoi(k)
		c, _ := strconv.Atoi(v)
		returnList = append(returnList, &chat.Unread{
			Id:    uint64(id),
			Count: uint32(c),
		})
	}

	return &chat.GetMessageNumberResp{Code: 200, Total: uint32(unReadTotal), List: returnList}, nil
}
