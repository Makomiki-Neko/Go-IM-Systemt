package groupmemberservicelogic

import (
	"context"

	"IMM/rpc/relation/internal/svc"
	"IMM/rpc/relation/relation"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProcessGroupJoinLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessGroupJoinLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessGroupJoinLogic {
	return &ProcessGroupJoinLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ProcessGroupJoinLogic) ProcessGroupJoin(in *relation.ProcessGroupJoinReq) (*relation.CommonResponse, error) {
	// todo: add your logic here and delete this line

	return &relation.CommonResponse{}, nil
}
