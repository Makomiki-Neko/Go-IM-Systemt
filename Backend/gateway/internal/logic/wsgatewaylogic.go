// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"IMM/gateway/internal/svc"
	"IMM/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WsGatewayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWsGatewayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WsGatewayLogic {
	return &WsGatewayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WsGatewayLogic) WsGateway(req *types.WsRequest) error {
	// todo: add your logic here and delete this line

	return nil
}
