package logic

import (
	"context"

	"IMM/common/models"
	"IMM/rpc/ai/ai"
	"IMM/rpc/ai/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type GetSessionListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSessionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSessionListLogic {
	return &GetSessionListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 拉取会话列表
func (l *GetSessionListLogic) GetSessionList(in *ai.GetSessionListReq) (*ai.GetSessionListResp, error) {
	// 起止记录
	if in.Page.Page < 1 {
		in.Page.Page = 1
	}
	starRecord := (in.Page.Page - 1) * in.Page.Size
	r, err := gorm.G[models.LlmSession](l.svcCtx.DB).Where("user_id = ?", in.UserId).Offset(int(starRecord)).Limit(int(in.Page.Size)).Find(l.ctx)
	if err != nil {
		return nil, err
	}

	c, err := gorm.G[models.LlmSession](l.svcCtx.DB).Where("user_id = ?", in.UserId).Count(l.ctx, "*")

	AiSessionList := make([]*ai.SessionInfo, 0, len(r))
	for _, item := range r {
		AiSessionList = append(AiSessionList, &ai.SessionInfo{
			SessionId: item.SessionID,
			AgentSet:  uint64(item.AgentSet),
			Tokens:    int64(item.Tokens),
			LastMsg:   item.LastMsgID,
		})
	}

	return &ai.GetSessionListResp{Base: &ai.CommonResponse{Code: 100, Msg: "OK"}, List: AiSessionList, Page: &ai.PageInfo{Total: int32(c), Page: in.Page.Page, Size: in.Page.Size}}, nil
}
