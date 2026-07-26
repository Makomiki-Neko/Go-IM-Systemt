package models

import (
	"time"

	"gorm.io/gorm"
)

// Ai	消息记录
type LlmMessage struct {
	MsgID     uint64 `gorm:"primaryKey;type:bigint unsigned;autoIncrement;comment:消息ID"`
	UserID    uint64 `gorm:"type:bigint unsigned;not null;index:idx_from_user;comment:用户ID"`
	SessionID uint64 `gorm:"type:bigint unsigned;not null;index:idx_user_session;comment:会话ID"`
	IsAiMsg   bool   `gorm:"type:tinyint;not null;default:False;comment:是否Ai生成"`
	MsgType   int8   `gorm:"type:tinyint;not null;default:1;comment:消息内容类型 1文本 2图片 3语音 4视频 5文件"`
	Content   string `gorm:"type:text;comment:消息内容"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 用户对智能体的配置
type AgentInfo struct {
	UserID   uint64 `gorm:"type:bigint unsigned;not null;index:idx_from_user;comment:用户ID"`
	Name     string `gorm:"type:text;comment:智能体名字"`
	Describe string `gorm:"type:text;comment:智能体人设描述"`
	Avatar   string `gorm:"type:text;comment:智能体头像链接"`
	gorm.Model
}

// 会话记录
type LlmSession struct {
	SessionID uint64 `gorm:"primarykey; type:bigint unsigned;not null;index:idx_user_session;comment:会话ID"`
	UserID    uint64 `gorm:"type:bigint unsigned;not null;index:idx_user_id;comment:发起用户ID"`
	AgentSet  uint   `gorm:"comment:绑定的智能体配置"`
	Tokens    uint64 `gorm:"type:bigint unsigned;comment:会话ID"`
	LastMsgID uint64 `gorm:"type:bigint unsigned;comment:最后已读消息ID"`
}

type SessionSummary struct {
	SessionID uint64 `gorm:"primarykey; type:bigint unsigned;not null;index:idx_user_session;comment:会话ID"`
	Content   string `gorm:"type:text;comment:总结的消息内容"`
	LastMsgID uint64 `gorm:"type:bigint unsigned;comment:最后总结的消息ID"`
}
