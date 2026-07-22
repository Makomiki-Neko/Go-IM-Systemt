package groupmemberservicelogic

import (
	"context"

	"IMM/rpc/relation/internal/svc"
	"IMM/rpc/relation/relation"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetGroupNickLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetGroupNickLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetGroupNickLogic {
	return &SetGroupNickLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetGroupNickLogic) SetGroupNick(in *relation.SetGroupNickReq) (*relation.CommonResponse, error) {
	// todo: add your logic here and delete this line

	return &relation.CommonResponse{}, nil
}
