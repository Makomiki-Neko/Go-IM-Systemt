package groupmemberservicelogic

import (
	"context"

	"IMM/rpc/relation/internal/svc"
	"IMM/rpc/relation/relation"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupJoinRequestsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupJoinRequestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupJoinRequestsLogic {
	return &GetGroupJoinRequestsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetGroupJoinRequestsLogic) GetGroupJoinRequests(in *relation.GetGroupJoinRequestsReq) (*relation.GetGroupJoinRequestsResp, error) {
	// todo: add your logic here and delete this line

	return &relation.GetGroupJoinRequestsResp{}, nil
}
