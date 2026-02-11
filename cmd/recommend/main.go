package main

import (
	"log"

	"shortvideo/internal/recommend/dao"
	"shortvideo/internal/recommend/handler"
	"shortvideo/internal/recommend/service"
	recommend "shortvideo/kitex_gen/recommend/recommendservice"
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

	//初始化推送dao
	actionRepo := dao.NewUserActionRepository(db)
	videoTagRepo := dao.NewVideoTagRepository(db)
	preferenceRepo := dao.NewUserPreferenceRepository(db)

	//初始化推送服务
	recommendService := service.NewRecommendService(
		actionRepo,
		videoTagRepo,
		preferenceRepo,
	)

	//初始化处理器
	recommendHandler := handler.NewRecommendService(recommendService)

	//创建ETCD注册器
	registry, err := registry_etcd.NewEtcdRegistry(cfg.Etcd.Endpoints)
	if err != nil {
		log.Fatalf("初始化ETCD注册器失败: %v", err)
	}

	//创建服务选项
	serverOpts := []server.Option{
		server.WithRegistry(registry),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: "recommend",
		}),
	}

	//创建服务器
	svr := recommend.NewServer(recommendHandler, serverOpts...)

	//启动服务器
	log.Printf("推荐服务启动，端口: %d", cfg.Ports.Recommend)
	err = svr.Run()
	if err != nil {
		logger.Error("Failed to start recommend server", logger.ErrorField(err))
		log.Println(err.Error())
	}

	logger.Info("Recommend server started successfully")
}
