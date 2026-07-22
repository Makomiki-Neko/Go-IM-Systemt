package logic

import (
	"context"
	"encoding/json"
	"errors"

	"IMM/common/models"
	"IMM/rpc/chat/chat"
	"IMM/rpc/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type GetHistoryChatMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetHistoryChatMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHistoryChatMessageLogic {
	return &GetHistoryChatMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 拉取历史消息
func (l *GetHistoryChatMessageLogic) GetHistoryChatMessage(in *chat.GetChatHistoryMessagesReq) (*chat.CommonResponse, error) {
	// 拉取与指定好友间发送的消息，以给定MsgID为界
	msgs, err := gorm.G[models.PrivateMessage](l.svcCtx.DB).
		Where(
			"(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
			in.FromUserId, in.UserId, // 朋友发给我的
			in.UserId, in.FromUserId, // 我发给朋友的
		).
		Where("msg_id < ?", in.StartMsgId).
		Order("msg_id Desc"). // 按消息ID排序
		Limit(int(in.Limit)).
		Find(l.ctx)

	if err != nil {
		return nil, errors.New("DB Failed, " + err.Error())
	}

	msgBytes, _ := json.Marshal(msgs)

	return &chat.CommonResponse{Code: 200, Data: msgBytes}, nil
}
