package groupmemberservicelogic

import (
	"context"

	"IMM/rpc/relation/internal/svc"
	"IMM/rpc/relation/relation"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApplyToJoinGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApplyToJoinGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyToJoinGroupLogic {
	return &ApplyToJoinGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ----- 群成员管理 -----
func (l *ApplyToJoinGroupLogic) ApplyToJoinGroup(in *relation.ApplyToJoinGroupReq) (*relation.CommonResponse, error) {
	// todo: add your logic here and delete this line

	return &relation.CommonResponse{}, nil
}
