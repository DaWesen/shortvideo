package main

import (
	"log"
	"net"

	"shortvideo/internal/interaction/dao"
	"shortvideo/internal/interaction/handler"
	"shortvideo/internal/interaction/service"
	videoDao "shortvideo/internal/video/dao"
	videoService "shortvideo/internal/video/service"
	"shortvideo/kitex_gen/interaction/interactionservice"
	"shortvideo/pkg/cache"
	"shortvideo/pkg/config"
	"shortvideo/pkg/database"
	"shortvideo/pkg/es"
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

	//初始化Elasticsearch
	esClient, err := es.NewESManager()
	if err != nil {
		log.Printf("初始化Elasticsearch客户端失败: %v，服务将继续运行", err)
	}

	//初始化视频DAO
	videoRepo := videoDao.NewVideoRepository(db)

	//初始化视频服务
	videoService := videoService.NewVideoService(videoRepo, minioClient, kafkaProducer, redisClient, esClient)

	//初始化互动DAO
	likeRepo := dao.NewLikeRepository(db)
	starRepo := dao.NewStarRepository(db)
	commentRepo := dao.NewCommentRepository(db)
	shareRepo := dao.NewShareRepository(db)
	statsRepo := dao.NewVideoInteractionStatsRepository(db)

	//初始化互动服务
	interactionService := service.NewInteractionService(likeRepo, starRepo, commentRepo, shareRepo, statsRepo, videoService, kafkaProducer)

	//初始化处理器
	interactionHandler := handler.NewInteractionService(interactionService)

	//创建ETCD注册器
	registry, err := registry_etcd.NewEtcdRegistry(cfg.Etcd.Endpoints)
	if err != nil {
		log.Fatalf("初始化ETCD注册器失败: %v", err)
	}

	//创建服务选项
	serverOpts := []server.Option{
		server.WithRegistry(registry),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: "interaction",
		}),
		server.WithServiceAddr(&net.TCPAddr{Port: cfg.Ports.Interaction}),
	}

	//创建服务
	svr := interactionservice.NewServer(interactionHandler, serverOpts...)

	//启动服务
	log.Printf("交互服务启动，端口: %d", cfg.Ports.Interaction)
	err = svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}
