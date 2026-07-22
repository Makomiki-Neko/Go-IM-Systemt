package groupmemberservicelogic

import (
	"context"

	"IMM/rpc/relation/internal/svc"
	"IMM/rpc/relation/relation"

	"github.com/zeromicro/go-zero/core/logx"
)

type InviteToGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInviteToGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InviteToGroupLogic {
	return &InviteToGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *InviteToGroupLogic) InviteToGroup(in *relation.InviteToGroupReq) (*relation.CommonResponse, error) {
	// todo: add your logic here and delete this line

	return &relation.CommonResponse{}, nil
}
