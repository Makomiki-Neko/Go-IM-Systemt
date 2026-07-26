package logic

import (
	"context"
	"encoding/json"
	"errors"

	"IMM/common/models"
	"IMM/rpc/ai/ai"
	"IMM/rpc/ai/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type GetHistoryAiMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetHistoryAiMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHistoryAiMessageLogic {
	return &GetHistoryAiMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 拉取历史消息, WS推送
func (l *GetHistoryAiMessageLogic) GetHistoryAiMessage(in *ai.GetChatHistoryMessagesFromAIReq) (*ai.CommonResponse, error) {
	// 复用Chat服务方法
	msgs, err := gorm.G[models.LlmMessage](l.svcCtx.DB).
		Where(
			"session_id = ? AND msg_id < ?",
			in.SessionId,
			in.StartMsgId,
		).
		Limit(int(in.Limit)).
		Find(l.ctx)

	if err != nil {
		return nil, errors.New("DB Failed, " + err.Error())
	}

	msgBytes, _ := json.Marshal(msgs)

	// 由GateWay网关推送
	return &ai.CommonResponse{Code: 200, Data: msgBytes}, nil
}
