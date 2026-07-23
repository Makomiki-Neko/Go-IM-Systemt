package models

import (
	"time"

	"gorm.io/gorm"
)

// Ai	消息记录
type LLMMessage struct {
	MsgID     uint64    `gorm:"primaryKey;type:bigint unsigned;autoIncrement;comment:消息ID"`
	UserID    uint64    `gorm:"type:bigint unsigned;not null;index:idx_from_user;comment:用户ID"`
	SessionID uint64    `gorm:"type:bigint unsigned;not null;index:idx_user_session;comment:会话ID"`
	IsAiMsg   bool      `gorm:"type:tinyint;not null;default:False;comment:是否Ai生成"`
	MsgType   int8      `gorm:"type:tinyint;not null;default:1;comment:消息内容类型"`
	Content   string    `gorm:"type:text;comment:消息内容"`
	SendTime  time.Time `gorm:"not null;index:idx_send_time;comment:消息发送时间"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
