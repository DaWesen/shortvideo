package database

import (
	"log"
	danmu_model "shortvideo/internal/danmu/model"
	interaction_model "shortvideo/internal/interaction/model"
	live_model "shortvideo/internal/live/model"
	message_model "shortvideo/internal/message/model"
	recommend_model "shortvideo/internal/recommend/model"
	social_model "shortvideo/internal/social/model"
	user_model "shortvideo/internal/user/model"
	video_model "shortvideo/internal/video/model"
	"shortvideo/pkg/config"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db     *gorm.DB
	dbOnce sync.Once
)

func Init() (*gorm.DB, error) {
	var err error
	dbOnce.Do(func() {
		cfg := config.Get()
		db, err = InitPostgres(cfg.Database.Postgres)
	})
	return db, err
}

func InitPostgres(postgresConfig config.PostgresConfig) (*gorm.DB, error) {
	dsn := postgresConfig.GetDSN()
	var logLevel logger.LogLevel
	logLevel = logger.Info

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		log.Printf("Failed to connect to database:%v", err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Failed to get sql.DB:%v", err)
		return nil, err
	}

	sqlDB.SetMaxIdleConns(postgresConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(postgresConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(postgresConfig.ConnMaxLifetime)
	log.Println("Database connected successfully")

	//自动迁移数据库表
	log.Println("开始自动迁移数据库表...")
	err = db.AutoMigrate(
		&user_model.User{},
		&video_model.Video{},
		&social_model.Follow{},
		&interaction_model.Comment{},
		&interaction_model.Like{},
		&interaction_model.Star{},
		&interaction_model.Share{},
		&interaction_model.VideoInteractionStats{},
		&message_model.Message{},
		&message_model.SystemNotification{},
		&live_model.LiveRoom{},
		&live_model.Gift{},
		&live_model.GiftRecord{},
		&live_model.RoomAdmin{},
		&live_model.LiveRecord{},
		&live_model.RoomViewer{},
		&danmu_model.Danmu{},
		&danmu_model.DanmuFilter{},
		&recommend_model.UserAction{},
		&recommend_model.VideoTag{},
		&recommend_model.UserPreference{},
	)
	if err != nil {
		log.Printf("数据库迁移失败: %v", err)
		return nil, err
	}
	log.Println("数据库迁移完成")

	return db, nil
}

func GetDB() *gorm.DB {
	if db == nil {
		panic("database not initialized")
	}
	return db
}
