package logic

import (
	"context"

	"IMM/common/models"
	"IMM/common/pkg"
	"IMM/rpc/user/internal/svc"
	"IMM/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdatePasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePasswordLogic {
	return &UpdatePasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdatePasswordLogic) UpdatePassword(in *user.UpdatePasswordReq) (*user.UpdatePasswordResp, error) {
	/*
		tokenAccount := l.ctx.Value("account").(string)
		tokenUid := l.ctx.Value("uid").(uint64)
		if in.Account != tokenAccount {
			return nil, errors.New("Request Parameter No Match Token")
		}
	*/
	encpw, err := pkg.EncryptPwd(in.Password)
	if err != nil {
		return nil, err
	}

	_, err = gorm.G[models.User](l.svcCtx.DB).Where("account = ?", in.Account).Updates(l.ctx, models.User{Password: encpw})
	if err != nil {
		return nil, err
	}
	return &user.UpdatePasswordResp{}, nil
}
