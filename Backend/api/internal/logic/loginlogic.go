// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"time"

	"IMM/api/internal/svc"
	"IMM/api/internal/types"
	"IMM/rpc/user/user"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	// todo: add your logic here and delete this line
	r, err := l.svcCtx.UserRpc.Login(l.ctx, &user.LoginReq{Account: req.Account, Password: req.Password, DeviceId: req.DeviceId, Platform: req.Platform})
	if err != nil {
		return nil, err
	}
	if r.Code == 100 {
		now := time.Now().Unix()
		accessExpire := l.svcCtx.Config.Auth.AccessExpire
		// 你可以将用户ID等自定义信息放入Claims中[reference:10][reference:11]
		claims := jwt.MapClaims{ // 附件信息
			"account": req.Account,
			"uid":     r.Uid,
			"exp":     now + accessExpire, // 到期时间
			"iat":     now,
		}
		token := jwt.New(jwt.SigningMethodHS256)
		token.Claims = claims
		// 使用配置中的密钥签名
		tokenString, errS := token.SignedString([]byte(l.svcCtx.Config.Auth.AccessSecret))
		if errS != nil {
			return nil, errS
		}
		r.AccessToken = tokenString
	}
	return &types.LoginResponse{Code: int(r.Code), Uid: r.Uid, AccessToken: r.AccessToken, RefreshToken: r.RefreshToken, Msg: "Login Success."}, nil
}
