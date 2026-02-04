package model

import (
	"time"
)

type Comment struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:评论ID"`
	UserID     int64     `gorm:"index;not null;comment:用户ID"`
	VideoID    int64     `gorm:"index;not null;comment:视频ID"`
	Content    string    `gorm:"type:text;not null;comment:评论内容"`
	CreateTime string    `gorm:"size:50;not null;comment:创建时间"`
	ReplyToID  int64     `gorm:"index;default:0;comment:回复的评论ID"`
	CreatedAt  time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (Comment) TableName() string {
	return "comments"
}

type Like struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;comment:点赞ID"`
	UserID    int64     `gorm:"index;not null;comment:用户ID"`
	VideoID   int64     `gorm:"index;not null;comment:视频ID"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (Like) TableName() string {
	return "likes"
}

type Star struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;comment:收藏ID"`
	UserID    int64     `gorm:"index;not null;comment:用户ID"`
	VideoID   int64     `gorm:"index;not null;comment:视频ID"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (Star) TableName() string {
	return "stars"
}

type Share struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;comment:分享ID"`
	UserID    int64     `gorm:"index;not null;comment:用户ID"`
	VideoID   int64     `gorm:"index;not null;comment:视频ID"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (Share) TableName() string {
	return "shares"
}

type VideoInteractionStats struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;comment:统计ID"`
	VideoID      int64     `gorm:"uniqueIndex;not null;comment:视频ID"`
	ViewCount    int64     `gorm:"default:0;comment:播放次数"`
	LikeCount    int64     `gorm:"default:0;comment:点赞数"`
	CommentCount int64     `gorm:"default:0;comment:评论数"`
	StarCount    int64     `gorm:"default:0;comment:收藏数"`
	ShareCount   int64     `gorm:"default:0;comment:分享数"`
	CreatedAt    time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (VideoInteractionStats) TableName() string {
	return "video_interaction_stats"
}
