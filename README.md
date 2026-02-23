# 短视频平台后端

## 项目概览

这是一个用 Go 语言写的短视频平台后端系统，基于微服务架构，使用了 CloudWeGo 生态的 Kitex 和 Hertz 框架。系统包含用户、视频、社交、评论、消息、直播、弹幕和推荐等功能模块，支持完整的短视频平台后端业务。

## 技术栈

- **Go 1.25.1**
- **CloudWeGo Kitex**（RPC框架）
- **CloudWeGo Hertz**（HTTP框架）
- **PostgreSQL**（数据库）
- **Redis**（缓存）
- **Kafka**（消息队列）
- **Etcd**（服务发现）
- **Elasticsearch**（搜索）
- **MinIO**（对象存储）
- **JWT**（认证）
- **Prometheus**（监控）

## 项目结构

```
shortvideo/
├── cmd/            # 各个服务的入口
│   ├── gateway/    # 网关服务（8080）
│   │   └── main.go
│   ├── user/       # 用户服务（8881）
│   │   └── main.go
│   ├── video/      # 视频服务（8882）
│   │   └── main.go
│   ├── social/     # 社交服务（8883）
│   │   └── main.go
│   ├── interaction/# 交互服务（8884）
│   │   └── main.go
│   ├── message/    # 消息服务（8885）
│   │   └── main.go
│   ├── live/       # 直播服务（8886）
│   │   └── main.go
│   ├── danmu/      # 弹幕服务（8887）
│   │   └── main.go
│   └── recommend/  # 推荐服务（8888）
│       └── main.go
├── configs/        # 配置文件
│   └── config.yaml
├── idl/            # 接口定义文件
│   ├── common.thrift
│   ├── user.thrift
│   ├── video.thrift
│   ├── social.thrift
│   ├── interaction.thrift
│   ├── message.thrift
│   ├── live.thrift
│   ├── danmu.thrift
│   └── recommend.thrift
├── internal/       # 内部实现
│   ├── gateway/    # 网关服务内部实现
│   │   ├── handler/    # HTTP和WebSocket处理
│   │   ├── middleware/ # 中间件
│   │   └── router/     # 路由配置
│   ├── user/       # 用户服务内部实现
│   │   ├── dao/        # 数据访问层
│   │   ├── model/      # 数据模型
│   │   ├── service/    # 业务逻辑层
│   │   └── handler/    # RPC处理
│   ├── video/      # 视频服务内部实现
│   │   ├── dao/        # 数据访问层
│   │   ├── model/      # 数据模型
│   │   ├── service/    # 业务逻辑层
│   │   └── handler/    # RPC处理
│   └── 其他服务/      # 其他服务内部实现（结构类似）
├── kitex_gen/      # Kitex生成的代码
│   ├── common/     # 公共类型
│   ├── user/       # 用户服务生成代码
│   ├── video/      # 视频服务生成代码
│   └── 其他服务/      # 其他服务生成代码
├── pkg/            # 公共工具包
│   ├── cache/      # 缓存工具（Redis）
│   ├── config/     # 配置工具
│   ├── database/   # 数据库工具（PostgreSQL）
│   ├── es/         # Elasticsearch工具
│   ├── jwt/        # JWT工具
│   ├── logger/     # 日志工具（Zap）
│   ├── mq/         # 消息队列工具（Kafka）
│   ├── prometheus/ # 监控工具
│   ├── registry/   # 服务注册工具（Etcd）
│   ├── storage/    # 存储工具（MinIO）
│   └── tracing/    # 追踪工具
├── script/         # 脚本文件
│   └── bootstrap.sh # 启动脚本
├── build.sh        # 构建脚本
├── go.mod          # Go模块文件
├── go.sum          # Go依赖校验文件
└── kitex_info.yaml # Kitex配置文件
```

## 主要功能

### 用户模块
- 注册登录
- 个人资料管理
- JWT认证

### 视频模块
- 视频上传存储
- 视频流和详情
- 视频搜索

### 社交模块
- 关注/取关用户
- 粉丝和关注列表

### 交互模块
- 点赞/取消点赞
- 评论功能
- 评论列表

### 消息模块
- 发送消息
- 消息列表

### 直播模块
- 开始/停止直播
- 直播列表

### 弹幕模块
- 发送弹幕
- 弹幕列表

### 推荐模块
- 视频推荐

## API接口

### 公开接口
- POST `/api/user/register` - 注册
- POST `/api/user/login` - 登录
- GET `/api/video/feed` - 视频流
- GET `/api/video/detail` - 视频详情
- GET `/api/search` - 搜索
- GET `/api/interaction/comments` - 评论列表
- GET `/api/danmu/list` - 弹幕列表
- GET `/api/live/list` - 直播列表

### 需要认证的接口
- GET `/api/auth/user/profile` - 用户资料
- PUT `/api/auth/user/update` - 更新资料
- POST `/api/auth/social/follow` - 关注
- POST `/api/auth/social/unfollow` - 取关
- GET `/api/auth/social/following` - 关注列表
- GET `/api/auth/social/follower` - 粉丝列表
- POST `/api/auth/interaction/like` - 点赞
- POST `/api/auth/interaction/unlike` - 取消点赞
- POST `/api/auth/interaction/comment` - 评论
- POST `/api/auth/message/send` - 发消息
- GET `/api/auth/message/list` - 消息列表
- POST `/api/auth/live/start` - 开始直播
- POST `/api/auth/live/stop` - 停止直播
- POST `/api/auth/danmu/send` - 发弹幕
- GET `/api/auth/recommend/videos` - 推荐视频

### WebSocket接口
- GET `/ws` - 实时通信（弹幕等）

### 接口文档

完整的API接口文档可通过Apifox访问：
- [Apifox接口文档](https://s.apifox.cn/94cd11a3-30a8-41e2-8ef1-b205bf3ccad1)

## 快速开始

### 环境准备
需要安装：
- Go 1.25.1+
- PostgreSQL
- Redis
- Kafka
- Etcd
- MinIO
- Elasticsearch

### 配置
修改 `configs/config.yaml` 文件，设置各个服务的地址和端口。

### 构建运行

```bash
# 构建网关服务
./build.sh

# 运行网关服务
./script/bootstrap.sh

# 构建其他服务（例如用户服务）
RUN_NAME="user" go build -o output/bin/${RUN_NAME} ./cmd/user

# 运行其他服务
./script/bootstrap.sh <运行目录> user
```

## 部署

### 本地开发
直接运行构建和启动脚本即可。


## 监控

- Prometheus 监控：`http://<服务地址>:9090/metrics`
- 日志使用 Zap 框架，可在配置文件中设置级别和格式。

## 安全

- JWT 认证
- 密码 bcrypt 加密
- 支持 HTTPS
- 请求限流
- 参数校验

## 性能优化

- Redis 缓存热点数据
- Kafka 处理异步任务
- 批量操作减少请求
- 数据库连接池
- 服务降级
