package logic

import (
	"context"

	"IMM/common/models"
	"IMM/rpc/ai/ai"
	"IMM/rpc/ai/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
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
	_, err := gorm.G[models.LlmSession](l.svcCtx.DB).Where("user_id = ? AND session_id = ? AND last_msg_id < ?", in.UserId, in.SessionId, in.MsgId).Updates(l.ctx, models.LlmSession{LastMsgID: in.MsgId})

	if err != nil {
		return &ai.CommonResponse{Code: 400, Msg: err.Error()}, nil
	}

	return &ai.CommonResponse{Code: 200, Msg: "OK"}, nil
}
