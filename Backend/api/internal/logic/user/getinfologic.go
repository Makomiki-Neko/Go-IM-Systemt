// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"IMM/api/internal/svc"
	"IMM/api/internal/types"
	"IMM/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInfoLogic {
	return &GetInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetInfoLogic) GetInfo(req *types.UserInfoRequest) (resp *types.UserInfoResponse, err error) {
	// todo: add your logic here and delete this line
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

	r, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &user.UserInfoReq{Account: req.Account, Uid: uint64(uid)})
	if err != nil {
		return nil, err
	}

	avatarURL := ""
	if r.Photo != "" {
		// 拼接 URL：base + "/" + fid
		avatarURL = fmt.Sprintf("%s/%s", l.svcCtx.Config.SeaweedFS.Filers[0], r.Photo)
	} else {
		// 没有头像时返回默认头像 URL
		avatarURL = fmt.Sprintf("%s/%s", l.svcCtx.Config.SeaweedFS.Filers[0], "buckets/my-bucket/avatars/default.jpg")
	}

	return &types.UserInfoResponse{Name: r.Name, Photo: avatarURL, Gender: r.Gender, Signature: r.Signature, Birthday: r.Birthday, CreatedAt: r.CreatedAt}, nil
}
