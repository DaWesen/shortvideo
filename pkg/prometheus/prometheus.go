package prometheus

import (
	"fmt"
	"log"
	"net/http"
	"shortvideo/pkg/config"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	promInstance *PrometheusManager
	promOnce     sync.Once
)

// 管理Prometheus监控
type PrometheusManager struct {
	config     *config.PrometheusConfig
	registry   *prometheus.Registry
	httpServer *http.Server
}

// 创建Prometheus管理器
func NewPrometheusManager() (*PrometheusManager, error) {
	var err error
	promOnce.Do(func() {
		cfg := config.Get().Prometheus
		registry := prometheus.NewRegistry()

		//注册默认指标
		registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
		registry.MustRegister(collectors.NewGoCollector())

		promInstance = &PrometheusManager{
			config:   &cfg,
			registry: registry,
		}

		//如果启用了Prometheus，则启动HTTP服务器
		if cfg.Enable {
			httpHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
			mux := http.NewServeMux()
			mux.Handle(cfg.Path, httpHandler)

			promInstance.httpServer = &http.Server{
				Addr:    fmt.Sprintf(":%d", cfg.Port),
				Handler: mux,
			}

			go func() {
				log.Printf("Prometheus监控已启动，端口: %d, 路径: %s", cfg.Port, cfg.Path)
				if err := promInstance.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Printf("Prometheus服务器启动失败: %v", err)
				}
			}()
		}
	})

	return promInstance, err
}
