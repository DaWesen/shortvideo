package model

import (
	"time"
)

type LiveRoom struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;comment:直播间ID"`
	HostID      int64     `gorm:"index;not null;comment:主播ID"`
	Title       string    `gorm:"size:200;not null;comment:标题"`
	CoverURL    string    `gorm:"size:500;comment:封面地址"`
	RtmpURL     string    `gorm:"size:500;comment:RTMP地址"`
	HlsURL      string    `gorm:"size:500;comment:HLS地址"`
	ViewerCount int64     `gorm:"default:0;comment:观众数"`
	IsLive      bool      `gorm:"default:false;comment:是否在直播"`
	CreateTime  string    `gorm:"size:50;not null;comment:创建时间"`
	CreatedAt   time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (LiveRoom) TableName() string {
	return "live_rooms"
}

type Gift struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;comment:礼物ID"`
	Name         string    `gorm:"size:100;not null;comment:礼物名称"`
	Price        int64     `gorm:"not null;comment:价格"`
	IconURL      string    `gorm:"size:500;comment:图标地址"`
	AnimationURL string    `gorm:"size:500;comment:动画地址"`
	CreatedAt    time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (Gift) TableName() string {
	return "gifts"
}

type GiftRecord struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:记录ID"`
	SenderID   int64     `gorm:"index;not null;comment:发送者ID"`
	RoomID     int64     `gorm:"index;not null;comment:直播间ID"`
	GiftID     int64     `gorm:"index;not null;comment:礼物ID"`
	Count      int32     `gorm:"default:1;comment:数量"`
	TotalPrice int64     `gorm:"comment:总价格"`
	CreatedAt  time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (GiftRecord) TableName() string {
	return "gift_records"
}

type RoomAdmin struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;comment:管理员ID"`
	RoomID    int64     `gorm:"index;not null;comment:直播间ID"`
	UserID    int64     `gorm:"index;not null;comment:用户ID"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (RoomAdmin) TableName() string {
	return "room_admins"
}

type LiveRecord struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;comment:录制ID"`
	RoomID    int64     `gorm:"index;not null;comment:直播间ID"`
	VideoURL  string    `gorm:"size:500;comment:视频地址"`
	StartTime time.Time `gorm:"comment:开始时间"`
	EndTime   time.Time `gorm:"comment:结束时间"`
	Duration  int64     `gorm:"default:0;comment:时长(秒)"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (LiveRecord) TableName() string {
	return "live_records"
}

type RoomViewer struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;comment:记录ID"`
	RoomID    int64      `gorm:"index;not null;comment:直播间ID"`
	UserID    int64      `gorm:"index;not null;comment:用户ID"`
	JoinTime  time.Time  `gorm:"comment:加入时间"`
	LeaveTime *time.Time `gorm:"comment:离开时间"`
	Duration  int64      `gorm:"default:0;comment:观看时长(秒)"`
	CreatedAt time.Time  `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime;comment:更新时间"`
}

func (RoomViewer) TableName() string {
	return "room_viewers"
}
