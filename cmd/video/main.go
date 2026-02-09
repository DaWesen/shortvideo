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

	//初始化视频DAO
	videoRepo := dao.NewVideoRepository(db)

	//初始化视频服务
	videoService := service.NewVideoService(videoRepo, minioClient, kafkaProducer, redisClient)

	//初始化处理器
	videoHandler := handler.NewVideoService(videoService)

	//创建服务
	svr := video.NewServer(videoHandler)

	//启动服务
	err = svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}
