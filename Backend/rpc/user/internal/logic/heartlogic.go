package logic

import (
	"context"
	"fmt"

	"IMM/rpc/user/internal/svc"
	"IMM/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HeartLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHeartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HeartLogic {
	return &HeartLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *HeartLogic) Heart(in *user.HeartReq) (*user.HeartResp, error) {
	key := fmt.Sprintf("user:refresh:%d:%s", in.Uid, in.Platform)

	storedToken, err := l.svcCtx.Redis.Get(key)

	if err != nil {
		return nil, err
	}

	// 判断 RefreshToken 是否一致
	if storedToken != in.RefreshToken {
		// 比对失败：要么被踢下线，要么 Token 伪造
		return nil, status.Error(codes.Unauthenticated, "账号已在其他设备登录或会话失效")
	}

	// 刷新 Token 续期
	l.svcCtx.Redis.Expire(key, 259200)
	// 在线状态续期
	statusKey := fmt.Sprintf("user:status:%d", in.Uid)
	l.svcCtx.Redis.Setex(statusKey, "Online", 60)

	return &user.HeartResp{Code: 100}, nil
}
