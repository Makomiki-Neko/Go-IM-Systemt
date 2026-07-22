package logic

import (
	"context"
	"time"

	"IMM/common/models"
	"IMM/rpc/user/internal/svc"
	"IMM/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoLogic) GetUserInfo(in *user.UserInfoReq) (*user.UserInfoResp, error) {
	// todo: add your logic here and delete this line
	/*
		tokenAccount := l.ctx.Value("account").(string)
		tokenUid := l.ctx.Value("uid").(uint64)
			if in.Account != tokenAccount {
				return nil, errors.New("Request Parameter No Match Token")
			}
	*/
	r, err := gorm.G[models.UserInfo](l.svcCtx.DB).Where("user_id = ?", in.Uid).First(l.ctx)
	if err != nil {
		return nil, err
	}
	uinfo := user.UserInfoResp{
		Name:       r.Name,
		Photo:      r.Photo,
		Gender:     r.Gender,
		Birthday:   unixOrZero(r.Birthday), // 避免为空时空指针异常
		CreatedAt:  r.CreatedAt.Unix(),     // 此字段非空
		Signature:  r.Signature,
		LastOnline: unixOrZero(r.LastOnline),
	}
	return &uinfo, nil
}

func unixOrZero(t *time.Time) int64 {
	if t == nil {
		return 0
	}
	return t.Unix()
}
