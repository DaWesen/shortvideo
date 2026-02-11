package handler

import (
	"shortvideo/kitex_gen/danmu/danmuservice"
	"shortvideo/kitex_gen/interaction/interactionservice"
	"shortvideo/kitex_gen/live/liveservice"
	"shortvideo/kitex_gen/message/messageservice"
	"shortvideo/kitex_gen/recommend/recommendservice"
	"shortvideo/kitex_gen/social/socialservice"
	"shortvideo/kitex_gen/user/userservice"
	"shortvideo/kitex_gen/video/videoservice"
	"shortvideo/pkg/config"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/transport"
	registry_etcd "github.com/kitex-contrib/registry-etcd"
)

// ServiceClients 包含所有微服务的客户端
type ServiceClients struct {
	UserClient        userservice.Client
	VideoClient       videoservice.Client
	SocialClient      socialservice.Client
	InteractionClient interactionservice.Client
	MessageClient     messageservice.Client
	LiveClient        liveservice.Client
	DanmuClient       danmuservice.Client
	RecommendClient   recommendservice.Client
}

// 初始化所有服务客户端
func InitServiceClients() (*ServiceClients, error) {
	//加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	//创建ETCD解析器
	resolver, err := registry_etcd.NewEtcdResolver(cfg.Etcd.Endpoints)
	if err != nil {
		return nil, err
	}

	//通用客户端选项
	commonOpts := []client.Option{
		client.WithTransportProtocol(transport.TTHeader),
		client.WithResolver(resolver),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: "gateway",
		}),
	}

	//初始化用户服务客户端
	userClient, err := userservice.NewClient(
		"user",
		commonOpts...,
	)
	if err != nil {
		return nil, err
	}

	//初始化视频服务客户端
	videoClient, err := videoservice.NewClient(
		"video",
		commonOpts...,
	)
	if err != nil {
		return nil, err
	}

	//初始化社交服务客户端
	socialClient, err := socialservice.NewClient(
		"social",
		commonOpts...,
	)
	if err != nil {
		return nil, err
	}

	//初始化交互服务客户端
	interactionClient, err := interactionservice.NewClient(
		"interaction",
		commonOpts...,
	)
	if err != nil {
		return nil, err
	}

	//初始化消息服务客户端
	messageClient, err := messageservice.NewClient(
		"message",
		commonOpts...,
	)
	if err != nil {
		return nil, err
	}

	//初始化直播服务客户端
	liveClient, err := liveservice.NewClient(
		"live",
		commonOpts...,
	)
	if err != nil {
		return nil, err
	}

	//初始化弹幕服务客户端
	danmuClient, err := danmuservice.NewClient(
		"danmu",
		commonOpts...,
	)
	if err != nil {
		return nil, err
	}

	//初始化推荐服务客户端
	recommendClient, err := recommendservice.NewClient(
		"recommend",
		commonOpts...,
	)
	if err != nil {
		return nil, err
	}

	return &ServiceClients{
		UserClient:        userClient,
		VideoClient:       videoClient,
		SocialClient:      socialClient,
		InteractionClient: interactionClient,
		MessageClient:     messageClient,
		LiveClient:        liveClient,
		DanmuClient:       danmuClient,
		RecommendClient:   recommendClient,
	}, nil
}
