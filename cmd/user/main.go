package main

import (
	"log"
	"shortvideo/internal/user/dao"
	"shortvideo/internal/user/handler"
	"shortvideo/internal/user/service"
	user "shortvideo/kitex_gen/user/userservice"
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
	userRepo := dao.NewUserRepository(db)

	//初始化用户服务
	userService := service.NewUserService(userRepo, jwtManager, minioClient, kafkaProducer, redisClient)

	//初始化处理器
	userHandler := handler.NewUserService(userService)

	//创建ETCD注册器
	registry, err := registry_etcd.NewEtcdRegistry(cfg.Etcd.Endpoints)
	if err != nil {
		log.Fatalf("初始化ETCD注册器失败: %v", err)
	}

	//创建服务选项
	serverOpts := []server.Option{
		server.WithRegistry(registry),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: "user",
		}),
	}

	//创建服务
	svr := user.NewServer(userHandler, serverOpts...)

	//启动服务
	log.Printf("用户服务启动，端口: %d", cfg.Ports.User)
	err = svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}
