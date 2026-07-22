package logic

import (
	"context"

	"IMM/rpc/ai/ai"
	"IMM/rpc/ai/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAiMsgLastReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAiMsgLastReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAiMsgLastReadLogic {
	return &UpdateAiMsgLastReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新用户最后已读信息
func (l *UpdateAiMsgLastReadLogic) UpdateAiMsgLastRead(in *ai.UpdateLastReadMsgRep) (*ai.CommonResponse, error) {
	// todo: add your logic here and delete this line

	return &ai.CommonResponse{}, nil
}
