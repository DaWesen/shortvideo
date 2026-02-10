package main

import (
	"log"

	"shortvideo/internal/interaction/dao"
	"shortvideo/internal/interaction/handler"
	"shortvideo/internal/interaction/service"
	videoDao "shortvideo/internal/video/dao"
	videoService "shortvideo/internal/video/service"
	"shortvideo/kitex_gen/interaction/interactionservice"
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
	videoRepo := videoDao.NewVideoRepository(db)

	//初始化视频服务
	videoService := videoService.NewVideoService(videoRepo, minioClient, kafkaProducer, redisClient)

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

	//创建服务
	svr := interactionservice.NewServer(interactionHandler)

	//启动服务
	err = svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}
