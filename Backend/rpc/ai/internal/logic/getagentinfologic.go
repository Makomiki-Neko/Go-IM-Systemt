package logic

import (
	"context"

	"IMM/rpc/ai/ai"
	"IMM/rpc/ai/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAgentInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAgentInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAgentInfoLogic {
	return &GetAgentInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取智能体信息
func (l *GetAgentInfoLogic) GetAgentInfo(in *ai.GetAgentInfoReq) (*ai.GetAgentInfoResp, error) {
	// todo: add your logic here and delete this line

	return &ai.GetAgentInfoResp{}, nil
}
