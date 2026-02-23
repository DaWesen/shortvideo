package main

import (
	"log"
	"net"

	"shortvideo/internal/live/dao"
	"shortvideo/internal/live/handler"
	"shortvideo/internal/live/service"
	live "shortvideo/kitex_gen/live/liveservice"
	"shortvideo/pkg/config"
	"shortvideo/pkg/database"
	"shortvideo/pkg/es"
	"shortvideo/pkg/logger"
	"shortvideo/pkg/prometheus"
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

	//初始化数据库连接
	db, err := database.Init()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	//初始化直播相关dao
	roomRepo := dao.NewLiveRoomRepository(db)
	giftRepo := dao.NewGiftRepository(db)
	giftRecordRepo := dao.NewGiftRecordRepository(db)
	roomAdminRepo := dao.NewRoomAdminRepository(db)
	liveRecordRepo := dao.NewLiveRecordRepository(db)
	roomViewerRepo := dao.NewRoomViewerRepository(db)

	//初始化Elasticsearch
	esClient, err := es.NewESManager()
	if err != nil {
		log.Printf("初始化Elasticsearch客户端失败: %v，服务将继续运行", err)
	} else {
		//创建直播间索引
		liveMapping := es.GenerateLiveMapping()
		if err := esClient.CreateIndex("lives", liveMapping); err != nil {
			log.Printf("创建直播间索引失败: %v", err)
		}
	}

	//初始化Prometheus监控
	_, err = prometheus.NewPrometheusManager(cfg.Prometheus.LivePort)
	if err != nil {
		log.Printf("初始化Prometheus失败: %v，服务将继续运行", err)
	}

	//初始化分布式链路追踪
	_, err = tracing.NewTracingManager()
	if err != nil {
		log.Printf("初始化Tracing失败: %v，服务将继续运行", err)
	}

	//初始化直播服务
	liveService := service.NewLiveService(
		roomRepo,
		giftRepo,
		giftRecordRepo,
		roomAdminRepo,
		liveRecordRepo,
		roomViewerRepo,
		esClient,
	)

	//初始化处理器
	liveHandler := handler.NewLiveService(liveService)

	//创建ETCD注册器
	registry, err := registry_etcd.NewEtcdRegistry(cfg.Etcd.Endpoints)
	if err != nil {
		log.Fatalf("初始化ETCD注册器失败: %v", err)
	}

	//创建服务选项
	serverOpts := []server.Option{
		server.WithRegistry(registry),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: "live",
		}),
		server.WithServiceAddr(&net.TCPAddr{Port: cfg.Ports.Live}),
	}

	//创建服务器
	svr := live.NewServer(liveHandler, serverOpts...)

	//启动服务器
	log.Printf("直播服务启动，端口: %d", cfg.Ports.Live)
	err = svr.Run()
	if err != nil {
		logger.Error("Failed to start live server", logger.ErrorField(err))
		log.Println(err.Error())
	}

	logger.Info("Live server started successfully")
}
