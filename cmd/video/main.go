package main

import (
	"log"
	"shortvideo/internal/video/dao"
	"shortvideo/internal/video/handler"
	"shortvideo/internal/video/service"
	video "shortvideo/kitex_gen/video/videoservice"
	"shortvideo/pkg/cache"
	"shortvideo/pkg/config"
	"shortvideo/pkg/database"
	"shortvideo/pkg/es"
	"shortvideo/pkg/mq"
	"shortvideo/pkg/prometheus"
	"shortvideo/pkg/storage"
	"shortvideo/pkg/tracing"

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
	} else {
		//创建视频索引
		videoMapping := es.GenerateVideoMapping()
		if err := esClient.CreateIndex("videos", videoMapping); err != nil {
			log.Printf("创建视频索引失败: %v", err)
		}
	}

	//初始化Prometheus监控
	_, err = prometheus.NewPrometheusManager()
	if err != nil {
		log.Printf("初始化Prometheus失败: %v，服务将继续运行", err)
	}

	//初始化分布式链路追踪
	_, err = tracing.NewTracingManager()
	if err != nil {
		log.Printf("初始化Tracing失败: %v，服务将继续运行", err)
	}

	//初始化视频DAO
	videoRepo := dao.NewVideoRepository(db)

	//初始化视频服务
	videoService := service.NewVideoService(videoRepo, minioClient, kafkaProducer, redisClient, esClient)

	//初始化处理器
	videoHandler := handler.NewVideoService(videoService)

	//创建ETCD注册器
	registry, err := registry_etcd.NewEtcdRegistry(cfg.Etcd.Endpoints)
	if err != nil {
		log.Fatalf("初始化ETCD注册器失败: %v", err)
	}

	//创建服务选项
	serverOpts := []server.Option{
		server.WithRegistry(registry),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: "video",
		}),
	}

	//创建服务
	svr := video.NewServer(videoHandler, serverOpts...)

	//启动服务
	log.Printf("视频服务启动，端口: %d", cfg.Ports.Video)
	err = svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}
