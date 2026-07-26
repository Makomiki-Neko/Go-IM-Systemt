package logic

import (
	"context"
	"errors"
	"fmt"

	"IMM/common/models"
	"IMM/rpc/ai/ai"
	"IMM/rpc/ai/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
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

	r, err := gorm.G[models.AgentInfo](l.svcCtx.DB).Where("user_id = ?", in.UserId).First(l.ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("Not Found Agent Info.")
		}
		return nil, err
	}

	return &ai.GetAgentInfoResp{AgentId: uint64(r.ID), AgentName: r.Name, AgentPrompt: r.Describe, AgentAvatar: r.Avatar}, nil
}
