package models

import (
	"time"

	"gorm.io/gorm"
)

// 用户关系表模型
type UserFriend struct {
	UserID    uint64 `gorm:"type:bigint unsigned;not null;uniqueIndex:uk_user_friend,priority:1"`
	FriendID  uint64 `gorm:"type:bigint unsigned;not null;uniqueIndex:uk_user_friend,priority:2"`
	Remark    string `gorm:"type:varchar(50);comment:好友备注"`
	Status    int8   `gorm:"type:tinyint;default:0;comment:0申请中 1已通过 2已拒绝 3已删除"`
	LastMsgID uint64 `gorm:"type:bigint unsigned;comment:最后已读消息ID"`
	gorm.Model
}

// 群组
type GroupInfo struct {
	GroupID     uint64 `gorm:"primaryKey;type:bigint unsigned;autoIncrement:false"`
	Name        string `gorm:"type:varchar(50);not null"`
	OwnerID     uint64 `gorm:"type:bigint unsigned;not null;index"`
	Avatar      string `gorm:"type:varchar(255)"`
	Notice      string `gorm:"type:varchar(255)"`
	MemberCount int    `gorm:"type:int;default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type GroupMember struct {
	GroupID   uint64 `gorm:"type:bigint unsigned;not null;uniqueIndex:uk_group_user,priority:1"`
	UserID    uint64 `gorm:"type:bigint unsigned;not null;uniqueIndex:uk_group_user,priority:2;index"`
	Role      int8   `gorm:"type:tinyint;default:0;comment:0普通 1管理员 2群主"`
	GroupNick string `gorm:"type:varchar(50)"`
	Status    int8   `gorm:"type:tinyint;default:0;comment:0申请中 1已通过 2已拒绝 3已删除"`
	LastMsgID uint64 `gorm:"type:bigint unsigned;comment:最后已读消息ID"`
	gorm.Model
}
