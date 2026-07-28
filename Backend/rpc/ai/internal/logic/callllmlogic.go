package logic

import (
	"context"
	"encoding/json"

	"IMM/common/models"
	"IMM/rpc/ai/ai"
	"IMM/rpc/ai/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type CallLlmLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCallLlmLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CallLlmLogic {
	return &CallLlmLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 调用 LLM 服务
func (l *CallLlmLogic) CallLlm(in *ai.CallLlmReq) (*ai.CommonResponse, error) {
	var req svc.ChatCompletionRequest
	err := json.Unmarshal(in.Req, &req)
	if err != nil {
		logx.Error("LLM RPC Server Umarshal Failed, ", err.Error())
		return nil, err
	}
	req.Model = l.svcCtx.Config.LLM.Model
	req.Temperature = l.svcCtx.Config.LLM.Temperature
	req.TopP = l.svcCtx.Config.LLM.TopP
	req.MaxTokens = l.svcCtx.Config.LLM.MaxTokens
	req.Stream = false

	r, e := l.svcCtx.LLM.CreateChatCompletion(context.Background(), req)
	if e != nil {
		logx.Error("LLM Server Failed, ", e.Error())
		return nil, e
	}

	// 记录入库
	aiMsgRecord := models.LlmMessage{
		UserID:    in.UserId,
		SessionID: in.SessionId,
		IsAiMsg:   true,
		MsgType:   1,
		Content:   r.Choices[0].Message.Content,
	}
	err = gorm.G[models.LlmMessage](l.svcCtx.DB).Create(l.ctx, &aiMsgRecord)
	_, err = gorm.G[models.LlmSession](l.svcCtx.DB).Where("session_id = ?", in.SessionId).Updates(l.ctx, models.LlmSession{Tokens: uint64(r.Usage.TotalTokens)})

	if err != nil {
		logx.Error("DB Failed, ", err.Error())
		return nil, err
	}

	//
	data, err := json.Marshal(r.Choices[0].Message.Content)
	if err != nil {
		logx.Error("Json Marshal Failed, ", err.Error())
		return nil, err
	}

	return &ai.CommonResponse{Code: 200, Data: data}, nil
}
