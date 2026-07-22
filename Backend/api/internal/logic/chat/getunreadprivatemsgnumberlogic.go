// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"context"
	"encoding/json"
	"errors"

	"IMM/api/internal/svc"
	"IMM/api/internal/types"
	"IMM/rpc/chat/chat"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUnreadPrivateMsgNumberLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUnreadPrivateMsgNumberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUnreadPrivateMsgNumberLogic {
	return &GetUnreadPrivateMsgNumberLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUnreadPrivateMsgNumberLogic) GetUnreadPrivateMsgNumber() (resp *types.GetUnreadNumberResp, err error) {
	// Get Uid From JWT
	uidVal := l.ctx.Value("uid")
	if uidVal == nil {
		return nil, errors.New("uid not found in context")
	}
	uidNum, ok := uidVal.(json.Number)
	if !ok {
		return nil, errors.New("invalid uid type in context")
	}
	uid, err := uidNum.Int64()
	if err != nil {
		return nil, errors.New("invalid uid value")
	}

	r, err := l.svcCtx.ChatRPC.GetNewChatMessageNumber(l.ctx, &chat.GetChatNewMessagesNumberReq{
		UserId: uint64(uid),
	})

	returnList := make([]types.UnReadInfo, 0, len(r.List))
	for _, v := range r.List {
		returnList = append(returnList, types.UnReadInfo{
			Id:    v.Id,
			Count: int(v.Count),
		})
	}

	return &types.GetUnreadNumberResp{List: returnList}, nil
}
