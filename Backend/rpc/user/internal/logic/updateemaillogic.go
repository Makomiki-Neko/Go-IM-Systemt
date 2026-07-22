package logic

import (
	"context"
	"errors"

	"IMM/common/models"
	"IMM/rpc/user/internal/svc"
	"IMM/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateEmailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateEmailLogic {
	return &UpdateEmailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateEmailLogic) UpdateEmail(in *user.UpdateEmailReq) (*user.UpdateEmailResp, error) {
	// 查重
	_, err := gorm.G[models.User](l.svcCtx.DB).Where("Email = ?", in.Email).First(l.ctx)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return &user.UpdateEmailResp{Code: 400}, err
	}
	if err == nil {
		return &user.UpdateEmailResp{Code: 400}, errors.New("Failed to Update, Email Exist.")
	}
	err = nil
	// 更新
	_, err = gorm.G[models.User](l.svcCtx.DB).Where("account = ?", in.Account).Select("email").Updates(l.ctx, models.User{Email: in.Email})
	if err != nil {
		return &user.UpdateEmailResp{Code: 400}, err
	}
	return &user.UpdateEmailResp{Code: 100}, nil
}
