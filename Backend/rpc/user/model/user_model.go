package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表模型
type User struct {
	UID       uint64   `gorm:"primaryKey;type:bigint unsigned;autoIncrement:false"`
	Account   string   `gorm:"type:varchar(100);not null;index"`
	Email     string   `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password  string   `gorm:"type:varchar(100);not null"`
	Info      UserInfo `gorm:"foreignKey:UserID;references:UID"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type UserInfo struct {
	UserID     uint64 `gorm:"primaryKey;type:bigint unsigned;index"`
	Name       string `gorm:"type:varchar(100);not null"`
	Photo      string
	Gender     string
	Signature  string `gorm:"type:varchar(250)"`
	Birthday   *time.Time
	LastOnline *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}
