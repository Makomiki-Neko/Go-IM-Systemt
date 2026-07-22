package groupmemberservicelogic

import (
	"context"

	"IMM/rpc/relation/internal/svc"
	"IMM/rpc/relation/relation"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetGroupMemberRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetGroupMemberRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetGroupMemberRoleLogic {
	return &SetGroupMemberRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetGroupMemberRoleLogic) SetGroupMemberRole(in *relation.SetGroupMemberRoleReq) (*relation.CommonResponse, error) {
	// todo: add your logic here and delete this line

	return &relation.CommonResponse{}, nil
}
