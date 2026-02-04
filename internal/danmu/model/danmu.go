package model

import (
	"time"
)

type Danmu struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:弹幕ID"`
	UserID     int64     `gorm:"index;not null;comment:用户ID"`
	LiveID     int64     `gorm:"index;not null;comment:直播间ID"`
	Content    string    `gorm:"type:text;not null;comment:弹幕内容"`
	Color      string    `gorm:"size:20;default:'#FFFFFF';comment:颜色"`
	CreateTime string    `gorm:"size:50;not null;comment:创建时间"`
	CreatedAt  time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (Danmu) TableName() string {
	return "danmus"
}

type DanmuFilter struct {
	ID            int64     `gorm:"primaryKey;autoIncrement;comment:过滤ID"`
	UserID        int64     `gorm:"index;not null;comment:用户ID"`
	LiveID        int64     `gorm:"index;not null;comment:直播间ID"`
	Keywords      string    `gorm:"type:text;comment:关键词(JSON数组)"`
	HideAnonymous bool      `gorm:"default:false;comment:隐藏匿名弹幕"`
	HideLowLevel  bool      `gorm:"default:false;comment:隐藏低等级用户弹幕"`
	CreatedAt     time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (DanmuFilter) TableName() string {
	return "danmu_filters"
}

type DanmuStats struct {
	LiveID             int64            `json:"live_id"`
	TotalDanmuCount    int64            `json:"total_danmu_count"`
	ActiveUserCount    int64            `json:"active_user_count"`
	PeakDanmuPerMinute int64            `json:"peak_danmu_per_minute"`
	WordCloud          map[string]int64 `json:"word_cloud"`
}
