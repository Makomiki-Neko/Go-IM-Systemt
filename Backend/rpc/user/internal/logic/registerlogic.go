package logic

import (
	"context"
	"errors"

	"IMM/common/models"
	"IMM/common/pkg"
	snowflake "IMM/common/pkg"
	"IMM/rpc/user/internal/svc"
	"IMM/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *user.RegisterReq) (*user.RegisterResp, error) {
	// todo: add your logic here and delete this line
	_, errs := gorm.G[models.User](l.svcCtx.DB).Where("account = ? OR email = ?", in.Account, in.Email).First(l.ctx)

	// 已存在记录
	if errs == nil {
		return &user.RegisterResp{Code: 400, Msg: "Account Or Email Is Exist."}, errors.New("Account Or Email Already Exist.")
	}
	// 不存在或出错
	if !errors.Is(errs, gorm.ErrRecordNotFound) {
		return nil, errs
	}

	encpw, err := pkg.EncryptPwd(in.Password)
	if err != nil {
		return nil, err
	}
	newUid := snowflake.GenUint64()
	newUser := models.User{
		UID:      newUid,
		Account:  in.Account,
		Email:    in.Email,
		Password: encpw,
		Info:     models.UserInfo{UserID: newUid, Name: in.Account},
	}
	err = gorm.G[models.User](l.svcCtx.DB).Create(l.ctx, &newUser)
	if err != nil {
		return nil, err
	}
	return &user.RegisterResp{Code: 100, Msg: "Register Findish."}, nil
}
