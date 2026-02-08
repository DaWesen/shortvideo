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
	
	return db, nil
}

func GetDB() *gorm.DB {
	if db == nil {
		panic("database not initialized")
	}
	return db
}
