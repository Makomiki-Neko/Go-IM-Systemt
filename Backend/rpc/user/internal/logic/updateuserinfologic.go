package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"IMM/common/models"
	"IMM/rpc/user/internal/svc"
	"IMM/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserInfoLogic {
	return &UpdateUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserInfoLogic) UpdateUserInfo(in *user.InfoUpdateReq) (*user.InfoUpdateResp, error) {
	var t *time.Time // 默认为 nil
	if in.Birthday > 0 {
		tt := time.Unix(in.Birthday, 0)
		t = &tt
	}
	newUinfo := models.UserInfo{
		Name:      in.Name,
		Gender:    in.Gender,
		Photo:     in.Photo,
		Signature: in.Signature,
		Birthday:  t,
	}
	fmt.Printf("newUinfo: %v \n", newUinfo)
	_, err := gorm.G[models.UserInfo](l.svcCtx.DB).Where("user_id = ?", in.Uid).Updates(l.ctx, newUinfo)

	if err != nil {
		return &user.InfoUpdateResp{Code: 500}, errors.New("DataBase Update Error, " + err.Error())
	}
	return &user.InfoUpdateResp{Code: 100}, nil
}
