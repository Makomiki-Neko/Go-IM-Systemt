package logic

import (
	"context"

	"IMM/rpc/ai/ai"
	"IMM/rpc/ai/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
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
func (l *SendPrivateMessageLogic) SendPrivateMessage(in *ai.SendMessageToAiReq) (*ai.SendMessageResp, error) {
	// todo: add your logic here and delete this line

	return &ai.SendMessageResp{}, nil
}
