package registry

import (
	"fmt"

	registry_etcd "github.com/kitex-contrib/registry-etcd"
	"github.com/spf13/viper"
)

// 创建ETCD注册中心
func NewRegistry() (interface{}, error) {
	//从配置中获取ETCD信息
	endpoints := viper.GetStringSlice("etcd.endpoints")

	//创建ETCD注册中心
	registry, err := registry_etcd.NewEtcdRegistry(endpoints)
	if err != nil {
		return nil, fmt.Errorf("创建ETCD注册中心失败: %w", err)
	}

	return registry, nil
}

// 创建ETCD解析器
func NewResolver() (interface{}, error) {
	//从配置中获取ETCD信息
	endpoints := viper.GetStringSlice("etcd.endpoints")

	//创建ETCD解析器
	resolver, err := registry_etcd.NewEtcdResolver(endpoints)
	if err != nil {
		return nil, fmt.Errorf("创建ETCD解析器失败: %w", err)
	}

	return resolver, nil
}
