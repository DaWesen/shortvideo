package model

import (
	"time"
)

type UserAction struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:行为ID"`
	UserID     int64     `gorm:"index;not null;comment:用户ID"`
	ItemID     int64     `gorm:"index;not null;comment:项目ID"`
	ItemType   string    `gorm:"size:20;not null;comment:项目类型(video/user/live)"`
	ActionType string    `gorm:"size:20;not null;comment:行为类型(like/view/comment/share)"`
	Duration   int32     `gorm:"default:0;comment:观看时长(秒)"`
	Score      float64   `gorm:"default:0.0;comment:评分"`
	Timestamp  string    `gorm:"size:50;not null;comment:时间戳"`
	CreatedAt  time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (UserAction) TableName() string {
	return "user_actions"
}

type VideoTag struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;comment:标签ID"`
	VideoID   int64     `gorm:"index;not null;comment:视频ID"`
	TagName   string    `gorm:"size:50;index;not null;comment:标签名"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (VideoTag) TableName() string {
	return "video_tags"
}

type UserPreference struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;comment:偏好ID"`
	UserID    int64     `gorm:"index;not null;comment:用户ID"`
	TagName   string    `gorm:"size:50;index;not null;comment:标签名"`
	Weight    float64   `gorm:"default:0.0;comment:权重"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (UserPreference) TableName() string {
	return "user_preferences"
}
