package logic

import (
	"context"

	"IMM/rpc/chat/chat"
	"IMM/rpc/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUnreadGroupMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUnreadGroupMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUnreadGroupMessagesLogic {
	return &GetUnreadGroupMessagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUnreadGroupMessagesLogic) GetUnreadGroupMessages(in *chat.GetGroupUnreadMessagesReq) (*chat.CommonResponse, error) {
	// todo: add your logic here and delete this line

	return &chat.CommonResponse{}, nil
}
