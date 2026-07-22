package logic

import (
	"context"

	"IMM/rpc/ai/ai"
	"IMM/rpc/ai/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetAgentInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetAgentInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetAgentInfoLogic {
	return &SetAgentInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 自定义智能体
func (l *SetAgentInfoLogic) SetAgentInfo(in *ai.SetAgentInfoReq) (*ai.CommonResponse, error) {
	// todo: add your logic here and delete this line

	return &ai.CommonResponse{}, nil
}
