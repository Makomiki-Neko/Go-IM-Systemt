// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"IMM/api/internal/svc"
	"IMM/api/internal/types"
	"IMM/rpc/user/user"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

type HeartLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	httpReq *http.Request // 用于接收http请求头, 需同时修改handler里传递该参数
}

func NewHeartLogic(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *HeartLogic {
	return &HeartLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		httpReq: r,
	}
}

func (l *HeartLogic) Heart(req *types.HeartRequest) (*types.HeartResponse, error) {
	_, err := l.svcCtx.UserRpc.Heart(l.ctx, &user.HeartReq{
		Uid:          req.Uid,
		RefreshToken: req.RefreshToken,
		DeviceId:     req.DeviceId,
		Platform:     req.Platform,
	})
	if err != nil {
		return nil, err
	}

	// AccessToken的重签发
	// 读取 Authorization Header
	authHeader := l.httpReq.Header.Get("Authorization")

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	if tokenString != "" {
		parser := jwt.Parser{UseJSONNumber: true}
		claims := jwt.MapClaims{}

		_, err := parser.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(l.svcCtx.Config.Auth.AccessSecret), nil
		})
		if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
			fmt.Printf("Error is %v \n", err)
			return &types.HeartResponse{Code: 412}, errors.New("Invalid Token Signature")
		}

		// 提取附加信息
		uidNum, ok := claims["uid"].(json.Number)
		uid, err := uidNum.Int64()
		uidN := uint64(uid)

		if uidN != req.Uid {
			return &types.HeartResponse{Code: 412}, errors.New("Invalid Uid in Token")
		}

		expNum, ok := claims["exp"].(json.Number)
		exp, err := expNum.Int64()
		now := time.Now().Unix()

		if !ok {
			return &types.HeartResponse{Code: 412}, errors.New("Invalid Token Claims Type")
		}
		if err != nil {
			return &types.HeartResponse{Code: 412}, errors.New("Claims Values Transformer Error")
		}

		// 判断剩余时间
		needRefresh := (exp - now) < l.svcCtx.Config.Auth.RefreshThreshold
		if needRefresh {
			claims := jwt.MapClaims{ // 附件信息
				"account": req.Uid,
				"uid":     uidN,
				"exp":     now + l.svcCtx.Config.Auth.AccessExpire, // 到期时间
				"iat":     now,
			}
			token := jwt.New(jwt.SigningMethodHS256)
			token.Claims = claims
			// 使用配置中的密钥签名
			tokenString, errS := token.SignedString([]byte(l.svcCtx.Config.Auth.AccessSecret))
			if errS != nil {
				return &types.HeartResponse{Code: 500}, errS
			}
			// 刷新
			return &types.HeartResponse{Code: 200, AccessToken: tokenString}, nil
		}
		// 未到阈值不需要刷新
		return &types.HeartResponse{Code: 200}, nil
	}
	// 未传入AT
	return &types.HeartResponse{Code: 200}, nil
}
