package logic

import (
	"context"
	"errors"
	"fmt"

	"IMM/common/models"
	"IMM/common/pkg"
	"IMM/rpc/user/internal/svc"
	"IMM/rpc/user/user"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *user.LoginReq) (*user.LoginResp, error) {
	// 核对用户名
	userinfo, err := gorm.G[models.User](l.svcCtx.DB).Select("uid", "password").Where("account = ?", in.Account).First(l.ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("Account does not exist!")
	}
	if err != nil {
		return nil, err
	}

	// 核对密码
	if !pkg.CheckPwd(userinfo.Password, in.Password) {
		return nil, errors.New("Login Failed, Password Error!")
	}

	// 设置状态
	refreshToken := uuid.New().String()
	key := fmt.Sprintf("user:refresh:%d:%s", userinfo.UID, in.Platform)
	err = l.svcCtx.Redis.Setex(key, refreshToken, 259200) // 3 Day
	if err != nil {
		return nil, err
	}

	statusKey := fmt.Sprintf("user:status:%d", userinfo.UID)
	err = l.svcCtx.Redis.Setex(statusKey, "Online", 60) // 1 Min
	if err != nil {
		return nil, err
	}

	return &user.LoginResp{Code: 100, Uid: userinfo.UID, AccessToken: "Here will be fill Token At ApiGetaway", RefreshToken: refreshToken}, nil
}
