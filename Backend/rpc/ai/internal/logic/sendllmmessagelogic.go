package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"IMM/common/models"
	"IMM/rpc/ai/ai"
	"IMM/rpc/ai/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type SendLLMMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendLLMMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendLLMMessageLogic {
	return &SendLLMMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发送消息给LLM
func (l *SendLLMMessageLogic) SendLLMMessage(in *ai.SendMessageToAiReq) (*ai.SendMessageResp, error) {
	// * 暂时未对图片类型信息进行处理

	// 拼接上下文、提交给LLM队列
	// 从Redis缓存中读
	key := fmt.Sprintf("im.ai.messages.%d", in.SessionId)
	historyMsg, err := l.svcCtx.Redis.Lrange(key, 0, -1)
	if err != nil {
		// 仅通知
		logx.Info("Redis Failed:", err.Error())
		err = nil
	}
	// 构建模型请求体
	var chatReq svc.ChatCompletionRequest
	chatReq.Messages = []svc.ChatMessage{}
	// 解析redis记录为ChatMessage格式
	for _, m := range historyMsg {
		var singleMsg svc.ChatMessage
		err := json.Unmarshal([]byte(m), &singleMsg)
		if err != nil {
			logx.Info("Redis Data Json Unmarshal Failed: ", err.Error())
			err = nil
			// 只通知
			continue
		}
		chatReq.Messages = append(chatReq.Messages, singleMsg)
	}
	if historyMsg != nil {
		chatReq.Messages = append(chatReq.Messages, svc.ChatMessage{Role: "user", Content: in.Content}) // 插入用户消息
	}

	var Msgs []models.LlmMessage
	// redis 过期
	if historyMsg == nil {
		var last uint64
		summary, errS := gorm.G[models.SessionSummary](l.svcCtx.DB).Where("session_id = ?", in.SessionId).First(l.ctx)
		if errS != nil {
			logx.Info("Get Ai Session Summary Failed:", err.Error())
			last = 0
		} else {
			last = summary.LastMsgID
		}
		Msgs, err = gorm.G[models.LlmMessage](l.svcCtx.DB).Where("user_id = ? AND session_id = ? AND msg_id > ?", in.UserId, in.SessionId, last).Order("msg_id desc").Limit(20).Find(l.ctx)
		logx.Info("Redis No Record And DB Search Failed:", err.Error())
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		// 拼接历史消息
		toLlmMsgListFromDB := make([]svc.ChatMessage, len(Msgs)+2)
		if errS == nil {
			toLlmMsgListFromDB[0] = svc.ChatMessage{Role: "assistant", Content: "以下内容是过去你与用户间对话的摘要总结：" + summary.Content}
		}
		toLlmMsgListFromDB[len(Msgs)+1] = svc.ChatMessage{Role: "user", Content: in.Content} // 用户新信息
		for i, m := range Msgs {
			role := "user"
			if m.IsAiMsg {
				role = "assistant"
			}
			toLlmMsgListFromDB[len(Msgs)-i-2] = svc.ChatMessage{
				Role:    role,
				Content: m.Content,
			}
		}

		chatReq.Messages = toLlmMsgListFromDB
		// 加入Redis缓存
		for _, m := range toLlmMsgListFromDB {
			l.svcCtx.Redis.Rpush(fmt.Sprintf("im.ai.messages.%d", in.SessionId), m)
		}
	}

	// 记录到数据库
	useMsg := models.LlmMessage{
		UserID:    in.UserId,
		SessionID: in.SessionId,
		IsAiMsg:   false,
		MsgType:   int8(in.MsgType),
		Content:   in.Content,
	}
	err = gorm.G[models.LlmMessage](l.svcCtx.DB).Create(l.ctx, &useMsg)
	if err != nil {
		logx.Info("DB Create Failed:", err.Error())
		return nil, err
	}

	// 消息队列投递
	sendData, err := json.Marshal(chatReq)
	if err != nil {
		logx.Info("Json Marshal Failed:", err.Error())
		return nil, err
	}
	err = l.svcCtx.RabbitMQ.Publish(l.ctx, "im.ai.chat.request", sendData)
	if err != nil {
		logx.Info("MQ Publish Failed:", err.Error())
		return nil, err
	}

	return &ai.SendMessageResp{MsgId: useMsg.MsgID, SendTime: useMsg.UpdatedAt.Unix(), CommonResponse: &ai.CommonResponse{Code: 100, Msg: "OK"}}, nil
}
