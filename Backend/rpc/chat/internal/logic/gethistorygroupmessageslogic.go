package logic

import (
	"context"

	"IMM/rpc/chat/chat"
	"IMM/rpc/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHistoryGroupMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetHistoryGroupMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHistoryGroupMessagesLogic {
	return &GetHistoryGroupMessagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetHistoryGroupMessagesLogic) GetHistoryGroupMessages(in *chat.GetGroupHistoryMessagesReq) (*chat.CommonResponse, error) {
	// todo: add your logic here and delete this line

	return &chat.CommonResponse{}, nil
}
