package model

import (
	"time"
)

type Follow struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;comment:关注ID"`
	UserID       int64     `gorm:"index;not null;comment:用户ID"`
	TargetUserID int64     `gorm:"index;not null;comment:被关注用户ID"`
	CreatedAt    time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (Follow) TableName() string {
	return "follows"
}

type FollowStats struct {
	UserID        int64 `json:"user_id"`
	FollowCount   int64 `json:"follow_count"`
	FollowerCount int64 `json:"follower_count"`
	FriendCount   int64 `json:"friend_count"`
}
