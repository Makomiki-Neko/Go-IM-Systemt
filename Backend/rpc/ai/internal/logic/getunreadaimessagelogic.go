package logic

import (
	"context"

	"IMM/rpc/ai/ai"
	"IMM/rpc/ai/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUnreadAiMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUnreadAiMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUnreadAiMessageLogic {
	return &GetUnreadAiMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 拉取新消息，WS推送
func (l *GetUnreadAiMessageLogic) GetUnreadAiMessage(in *ai.GetChatUnreadMessagesFromAIReq) (*ai.CommonResponse, error) {
	// todo: add your logic here and delete this line

	return &ai.CommonResponse{}, nil
}
