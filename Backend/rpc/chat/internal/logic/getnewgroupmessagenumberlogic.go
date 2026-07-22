package logic

import (
	"context"

	"IMM/rpc/chat/chat"
	"IMM/rpc/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNewGroupMessageNumberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetNewGroupMessageNumberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNewGroupMessageNumberLogic {
	return &GetNewGroupMessageNumberLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetNewGroupMessageNumberLogic) GetNewGroupMessageNumber(in *chat.GetGroupNewMessagesNumberReq) (*chat.GetMessageNumberResp, error) {
	// todo: add your logic here and delete this line

	return &chat.GetMessageNumberResp{}, nil
}
