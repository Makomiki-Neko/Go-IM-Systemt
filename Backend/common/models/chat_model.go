package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// PrivateMessage 私聊消息记录
type PrivateMessage struct {
	MsgID      uint64         `gorm:"primaryKey;type:bigint unsigned;autoIncrement;comment:消息ID" json:"msg_id,string"`
	FromUserID uint64         `gorm:"type:bigint unsigned;not null;index:idx_from_user;comment:发送者ID" json:"from_user_id,string"`
	ToUserID   uint64         `gorm:"type:bigint unsigned;not null;index:idx_to_user;comment:接收者ID" json:"to_user_id,string"`
	MsgType    int8           `gorm:"type:tinyint;not null;default:1;comment:消息类型 1文本 2图片 3语音 4视频 5文件" json:"msg_type"`
	Content    string         `gorm:"type:text;comment:消息内容（文本或文件URL）" json:"content"`
	SendTime   time.Time      `gorm:"not null;index:idx_send_time;comment:消息发送时间" json:"send_time"`
	Status     int8           `gorm:"type:tinyint;not null;default:0;comment:消息状态 0已发送 1已送达 废弃 改用水位机制不再维护" json:"status"`
	UpdatedAt  time.Time      `json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (p PrivateMessage) MarshalJSON() ([]byte, error) {
	type Alias PrivateMessage
	return json.Marshal(&struct {
		SendTime int64 `json:"send_time"`
		*Alias
	}{
		SendTime: p.SendTime.Unix(),
		Alias:    (*Alias)(&p),
	})
}

// GroupMessage 群聊消息记录
type GroupMessage struct {
	MsgID      uint64    `gorm:"primaryKey;type:bigint unsigned;autoIncrement;comment:消息ID"`
	FromUserID uint64    `gorm:"type:bigint unsigned;not null;index:idx_from_user;comment:发送者ID"`
	GroupID    uint64    `gorm:"type:bigint unsigned;not null;index:idx_group_id;comment:群组ID"`
	MsgType    int8      `gorm:"type:tinyint;not null;default:1;comment:消息类型"`
	Content    string    `gorm:"type:text;comment:消息内容"`
	SendTime   time.Time `gorm:"not null;index:idx_send_time;comment:消息发送时间"`
	// 群聊消息可增加“@列表”等扩展字段，这里暂不添加
	//AtUserIDs string     `gorm:"type:varchar(255);comment:被@的用户ID列表，JSON数组"`
	//IsRevoked bool       `gorm:"default:false;comment:是否撤回"`
	//RevokedAt *time.Time `gorm:"comment:撤回时间"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
