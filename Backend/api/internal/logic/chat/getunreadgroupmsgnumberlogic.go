// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"context"

	"IMM/api/internal/svc"
	"IMM/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUnreadGroupMsgNumberLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUnreadGroupMsgNumberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUnreadGroupMsgNumberLogic {
	return &GetUnreadGroupMsgNumberLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUnreadGroupMsgNumberLogic) GetUnreadGroupMsgNumber() (resp *types.GetUnreadNumberResp, err error) {
	// todo: add your logic here and delete this line

	return
}
