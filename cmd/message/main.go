package main

import (
	"log"
	"shortvideo/internal/message/dao"
	"shortvideo/internal/message/handler"
	"shortvideo/internal/message/service"
	userDao "shortvideo/internal/user/dao"
	userService "shortvideo/internal/user/service"
	"shortvideo/kitex_gen/message/messageservice"
	"shortvideo/pkg/cache"
	"shortvideo/pkg/config"
	"shortvideo/pkg/database"
	"shortvideo/pkg/jwt"
	"shortvideo/pkg/mq"
	"shortvideo/pkg/storage"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	registry_etcd "github.com/kitex-contrib/registry-etcd"
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
	redisClient := cache.NewRedisCache()

	//初始化Kafka生产者
	kafkaProducer := mq.NewProducer()
	if kafkaProducer == nil {
		log.Fatalf("初始化Kafka生产者失败")
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

	//创建ETCD注册器
	registry, err := registry_etcd.NewEtcdRegistry(cfg.Etcd.Endpoints)
	if err != nil {
		log.Fatalf("初始化ETCD注册器失败: %v", err)
	}

	//创建服务选项
	serverOpts := []server.Option{
		server.WithRegistry(registry),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: "message",
		}),
	}

	//创建服务
	svr := messageservice.NewServer(messageHandler, serverOpts...)

	//启动服务
	log.Printf("消息服务启动，端口: %d", cfg.Ports.Message)
	err = svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}
