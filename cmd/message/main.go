package main

import (
	"log"

	"shortvideo/internal/message/dao"
	"shortvideo/internal/message/handler"
	"shortvideo/internal/message/service"
	userDao "shortvideo/internal/user/dao"
	userService "shortvideo/internal/user/service"
	messageservice "shortvideo/kitex_gen/message/messageservice"
	"shortvideo/pkg/cache"
	"shortvideo/pkg/config"
	"shortvideo/pkg/database"
	"shortvideo/pkg/jwt"
	"shortvideo/pkg/mq"
	"shortvideo/pkg/storage"
)

func main() {
	//初始化配置
	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	//初始化数据库
	db, err := database.InitPostgres(cfg.Database.Postgres)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	//初始化Redis
	redisClient, err := cache.InitRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("初始化Redis失败: %v", err)
	}

	//初始化Kafka生产者
	kafkaProducer, err := mq.InitProducer(cfg.Kafka.Brokers, cfg.Kafka.Version)
	if err != nil {
		log.Fatalf("初始化Kafka生产者失败: %v", err)
	}

	//初始化MinIO
	minioClient, err := storage.InitMinio(cfg.Minio)
	if err != nil {
		log.Fatalf("初始化MinIO失败: %v", err)
	}

	//初始化JWT管理器
	jwtManager := jwt.NewJWTManagerWithConfig(cfg.JWT.Secret, cfg.JWT.ExpireHours)

	//初始化用户DAO
	userRepo := userDao.NewUserRepository(db)

	//初始化用户服务
	userService := userService.NewUserService(userRepo, jwtManager, minioClient, kafkaProducer, redisClient)

	//初始化消息DAO
	messageRepo := dao.NewMessageRepository(db)
	notificationRepo := dao.NewNotificationRepository(db)

	//初始化消息服务
	messageService := service.NewMessageService(messageRepo, notificationRepo, userService, kafkaProducer)

	//初始化处理器
	messageHandler := handler.NewMessageService(messageService)

	//创建服务
	svr := messageservice.NewServer(messageHandler)

	//启动服务
	err = svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}
