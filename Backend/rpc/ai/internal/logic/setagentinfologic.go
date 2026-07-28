package logic

import (
	"context"

	"IMM/common/models"
	"IMM/rpc/ai/ai"
	"IMM/rpc/ai/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
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
	_, err := gorm.G[models.AgentInfo](l.svcCtx.DB).Where("user_id = ? AND id = ?", in.UserId, in.AgentId).Updates(l.ctx, models.AgentInfo{Name: in.AgentName, Describe: in.AgentPrompt, Avatar: in.AgentAvatar})

	if err != nil {
		return nil, err
	}

	return &ai.CommonResponse{Code: 200, Msg: "OK"}, nil
}
