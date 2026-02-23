package main

import (
	"log"
	"net"
	"shortvideo/internal/danmu/dao"
	"shortvideo/internal/danmu/handler"
	"shortvideo/internal/danmu/service"
	danmu "shortvideo/kitex_gen/danmu/danmuservice"
	"shortvideo/pkg/config"
	"shortvideo/pkg/database"
	"shortvideo/pkg/logger"

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

	//初始化弹幕相关dao
	danmuRepo := dao.NewDanmuRepository(db)
	danmuFilterRepo := dao.NewDanmuFilterRepository(db)

	//初始化弹幕服务
	danmuService := service.NewDanmuService(
		danmuRepo,
		danmuFilterRepo,
	)

	//初始化处理器
	danmuHandler := handler.NewDanmuService(danmuService)

	//创建ETCD注册器
	registry, err := registry_etcd.NewEtcdRegistry(cfg.Etcd.Endpoints)
	if err != nil {
		log.Fatalf("初始化ETCD注册器失败: %v", err)
	}

	//创建服务选项
	serverOpts := []server.Option{
		server.WithRegistry(registry),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: "danmu",
		}),
		server.WithServiceAddr(&net.TCPAddr{Port: cfg.Ports.Danmu}),
	}

	//创建服务器
	svr := danmu.NewServer(danmuHandler, serverOpts...)

	//启动服务器
	log.Printf("弹幕服务启动，端口: %d", cfg.Ports.Danmu)
	err = svr.Run()
	if err != nil {
		logger.Error("Failed to start danmu server", logger.ErrorField(err))
		log.Println(err.Error())
	}

	logger.Info("Danmu server started successfully")
}
