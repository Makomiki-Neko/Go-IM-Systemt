package logic

import (
	"context"

	"IMM/rpc/ai/ai"
	"IMM/rpc/ai/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHistoryAiMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetHistoryAiMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHistoryAiMessageLogic {
	return &GetHistoryAiMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 拉取历史消息, WS推送
func (l *GetHistoryAiMessageLogic) GetHistoryAiMessage(in *ai.GetChatHistoryMessagesFromAIReq) (*ai.CommonResponse, error) {
	// todo: add your logic here and delete this line

	return &ai.CommonResponse{}, nil
}
