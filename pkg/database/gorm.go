package database

import (
	"log"
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
		dsn := cfg.Database.Postgres.GetDSN()
		var logLevel logger.LogLevel
		if cfg.App.Env == "dev" {
			logLevel = logger.Info
		} else {
			logLevel = logger.Silent
		}
		gormConfig := &gorm.Config{
			Logger: logger.Default.LogMode(logLevel),
			NowFunc: func() time.Time {
				return time.Now().Local()
			},
		}
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err != nil {
			log.Fatalf("Failed to connect to database:%v", err)
			return
		}
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("Failed to get sql.DB:%v", err)
			return
		}
		sqlDB.SetMaxIdleConns(cfg.Database.Postgres.MaxIdleConns)
		sqlDB.SetMaxOpenConns(cfg.Database.Postgres.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(cfg.Database.Postgres.ConnMaxLifetime)
		log.Println("Database connected successfully")
	})
	return db, err
}

func GetDB() *gorm.DB {
	if db == nil {
		panic("database not initialized")
	}
	return db
}
