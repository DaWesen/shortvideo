package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shortvideo/internal/gateway/handler"
	"shortvideo/internal/gateway/middleware"
	"shortvideo/internal/gateway/router"
	"shortvideo/pkg/config"
	"shortvideo/pkg/prometheus"
	"shortvideo/pkg/tracing"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func main() {
	//加载配置
	log.Printf("加载配置...")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}
	log.Printf("配置加载成功：网关端口=%d", cfg.Ports.Gateway)

	//初始化配置实例
	log.Printf("初始化配置实例...")
	_, err = config.Init()
	if err != nil {
		log.Fatalf("配置实例初始化失败: %v", err)
	}
	log.Printf("配置实例初始化成功")

	//初始化Prometheus监控
	log.Printf("初始化Prometheus监控...")
	_, err = prometheus.NewPrometheusManager(cfg.Prometheus.GatewayPort)
	if err != nil {
		log.Printf("警告: Prometheus初始化失败: %v", err)
	} else {
		log.Printf("Prometheus监控初始化成功")
	}

	//初始化分布式链路追踪
	log.Printf("初始化分布式链路追踪...")
	_, err = tracing.NewTracingManager()
	if err != nil {
		log.Printf("警告: Tracing初始化失败: %v", err)
	} else {
		log.Printf("分布式链路追踪初始化成功")
	}

	//初始化服务客户端
	log.Printf("初始化服务客户端...")
	serviceClients, err := handler.InitServiceClients()
	if err != nil {
		log.Printf("警告: 服务客户端初始化失败: %v", err)
		log.Printf("继续以有限功能运行...")
	}
	log.Printf("服务客户端初始化完成")

	//初始化WebSocket管理器
	log.Printf("初始化WebSocket管理器...")
	handler.InitWSManager()
	log.Printf("WebSocket管理器初始化完成")

	//创建HTTP处理器
	log.Printf("创建HTTP处理器...")
	httpHandler := handler.NewHTTPHandler(serviceClients)
	log.Printf("HTTP处理器创建成功")

	//创建中间件
	log.Printf("创建中间件...")
	authMiddleware := middleware.NewAuthMiddleware(serviceClients.UserClient)
	corsMiddleware := middleware.NewCORSMiddleware()
	loggerMiddleware := middleware.NewLoggerMiddleware()
	recoveryMiddleware := middleware.NewRecoveryMiddleware()
	log.Printf("中间件创建成功")

	//创建Hertz服务器
	port := cfg.Ports.Gateway
	srv := server.New(
		server.WithHostPorts(fmt.Sprintf(":%d", port)),
	)
	log.Printf("Hertz服务器创建成功：端口=%d", port)

	//注册路由
	log.Printf("注册路由...")
	router.RegisterRoutes(srv, httpHandler, authMiddleware, corsMiddleware, loggerMiddleware, recoveryMiddleware)
	log.Printf("路由注册成功")

	//启动服务器
	go func() {
		log.Printf("网关服务器正在端口 %d 上启动", port)
		srv.Spin()
	}()

	//等待中断信号
	log.Printf("网关服务器已启动，等待中断信号...")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Printf("正在关闭网关服务器...")

	//关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("网关服务器强制关闭: %v", err)
	}

	log.Printf("网关服务器退出")
}
