package model

import (
	"time"
)

type Video struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;comment:视频ID"`
	AuthorID     int64     `gorm:"index;not null;comment:作者ID"`
	URL          string    `gorm:"type:varchar(500);not null;comment:视频地址"`
	CoverURL     string    `gorm:"type:varchar(500);comment:封面地址"`
	LikeCount    int64     `gorm:"default:0;comment:点赞数"`
	CommentCount int64     `gorm:"default:0;comment:评论数"`
	Title        string    `gorm:"size:200;not null;comment:标题"`
	PublishTime  int64     `gorm:"index;not null;comment:发布时间戳"`
	Description  string    `gorm:"type:text;comment:描述"`
	CreatedAt    time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (Video) TableName() string {
	return "videos"
}

type VideoStats struct {
	VideoID      int64 `json:"video_id"`
	ViewCount    int64 `json:"view_count"`
	LikeCount    int64 `json:"like_count"`
	CommentCount int64 `json:"comment_count"`
	ShareCount   int64 `json:"share_count"`
}
