package model

import (
	"time"
)

type User struct {
	ID            int64     `gorm:"primaryKey;autoIncrement;comment:用户ID"`
	Username      string    `gorm:"size:32;uniqueIndex;not null;comment:用户名"`
	Password      string    `gorm:"size:128;not null;comment:密码"`
	Avatar        string    `gorm:"size:255;default:'';comment:头像"`
	About         string    `gorm:"type:text;default:'';comment:个人简介"`
	FollowCount   int64     `gorm:"default:0;comment:关注数"`
	FollowerCount int64     `gorm:"default:0;comment:粉丝数"`
	CreatedAt     time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (User) TableName() string {
	return "users"
}

type UserStats struct {
	UserID             int64 `json:"user_id"`
	VideoCount         int64 `json:"video_count"`
	TotalLikesReceived int64 `json:"total_likes_received"`
	TotalComments      int64 `json:"total_comments"`
}
