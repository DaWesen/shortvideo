package model

import (
	"time"
)

type Message struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:消息ID"`
	ReceiveID  int64     `gorm:"index;not null;comment:接收者ID"`
	SendID     int64     `gorm:"index;not null;comment:发送者ID"`
	Content    string    `gorm:"type:text;not null;comment:消息内容"`
	CreateTime string    `gorm:"size:50;not null;comment:创建时间"`
	IsRead     bool      `gorm:"default:false;comment:是否已读"`
	CreatedAt  time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (Message) TableName() string {
	return "messages"
}

type SystemNotification struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:通知ID"`
	UserID     int64     `gorm:"index;not null;comment:用户ID"`
	Title      string    `gorm:"size:200;not null;comment:标题"`
	Content    string    `gorm:"type:text;not null;comment:内容"`
	Type       int32     `gorm:"not null;comment:类型"`
	RelatedID  int64     `gorm:"index;comment:相关ID"`
	IsRead     bool      `gorm:"default:false;comment:是否已读"`
	CreateTime string    `gorm:"size:50;not null;comment:创建时间"`
	CreatedAt  time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (SystemNotification) TableName() string {
	return "system_notifications"
}
