package friendservicelogic

import (
	"context"
	"encoding/json"
	"errors"

	"IMM/common/models"
	"IMM/common/types"
	"IMM/rpc/relation/internal/svc"
	"IMM/rpc/relation/relation"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type SendFriendRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendFriendRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendFriendRequestLogic {
	return &SendFriendRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ----- 发送好友请求 -----
func (l *SendFriendRequestLogic) SendFriendRequest(in *relation.SendFriendRequestReq) (*relation.SendFriendRequestResp, error) {
	// 存在判断
	r, err := gorm.G[models.UserFriend](l.svcCtx.DB).Where("user_id = ? AND friend_id = ?", in.UserId, in.FriendId).First(l.ctx)
	var friendReq models.UserFriend
	var recordID uint
	// 已发送这个好友请求
	if err == nil {
		switch r.Status {
		case 1: // 已经是好友
			return &relation.SendFriendRequestResp{Base: &relation.CommonResponse{Code: 201, Msg: "Already is Friend"}, RequestId: uint64(r.ID)}, nil
		case 3: // 删除再添加
			_, err = gorm.G[models.UserFriend](l.svcCtx.DB).Where("user_id = ? AND friend_id = ?", in.UserId, in.FriendId).Updates(l.ctx, models.UserFriend{Status: 0, Remark: in.Msg})
			if err != nil {
				return nil, errors.New("DB Create Record Failed; " + err.Error())
			}
			friendReq = models.UserFriend{
				UserID:   in.UserId,
				FriendID: in.FriendId,
				Remark:   in.Msg, // 临时使用Remark字段作为申请信息
				Status:   0,
			}
			recordID = r.ID
		default: // 已发送但未回应
			return &relation.SendFriendRequestResp{Base: &relation.CommonResponse{Code: 201, Msg: "Friend Req Has Been Send"}, RequestId: uint64(r.ID)}, nil
		}
	} else if errors.Is(gorm.ErrRecordNotFound, err) {
		// 新建好友请求
		// 记录ID会自动回填
		friendReq = models.UserFriend{
			UserID:   in.UserId,
			FriendID: in.FriendId,
			Remark:   in.Msg, // 临时使用Remark字段作为申请信息
			Status:   0,
		}
		err = gorm.G[models.UserFriend](l.svcCtx.DB).Create(l.ctx, &friendReq)
		if err != nil {
			return nil, errors.New("DB Create Record Failed; " + err.Error())
		}
		recordID = friendReq.ID
	} else {
		return nil, errors.New("DB Failed; " + err.Error())
	}

	// 加入消息队列
	msg, err := json.Marshal(map[string]interface{}{
		"user_id":   in.UserId,
		"friend_id": in.FriendId,
		"msg":       in.Msg,
		"record_id": recordID,
	})
	packMsg := types.MqMsg{
		Uid:  in.FriendId,
		Type: "event.friendApply",
		Data: json.RawMessage(msg),
	}
	if err != nil {
		logx.Errorf("Friend Apply Msg Format Failed: %v", err)
	}
	sendMsg, err := json.Marshal(packMsg)
	routKey := "im.gateway.push.event.relation.friend.apply"
	err = l.svcCtx.RabbitMQ.Publish(l.ctx, routKey, sendMsg)
	if err != nil {
		// 注意：MQ 投递失败建议打印告警，不要直接阻断主流程（可根据业务要求调整）
		logx.Errorf("Friend Apply Msg Send Failed: %v", err)
	}
	return &relation.SendFriendRequestResp{Base: &relation.CommonResponse{Code: 200, Msg: "Friend Req Send Success."}, RequestId: uint64(recordID)}, nil
}
