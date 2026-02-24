package config

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/spf13/viper"
)

var (
	configInstance *Config
	configOnce     sync.Once
)

// Config 统一配置结构
type Config struct {
	App           AppConfig           `mapstructure:"app"`
	Ports         PortsConfig         `mapstructure:"ports"`
	Services      ServicesConfig      `mapstructure:"services"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Kafka         KafkaConfig         `mapstructure:"kafka"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
	Minio         MinioConfig         `mapstructure:"minio"`
	Log           LogConfig           `mapstructure:"log"`
	JWT           JWTConfig           `mapstructure:"jwt"`
	Prometheus    PrometheusConfig    `mapstructure:"prometheus"`
	Tracing       TracingConfig       `mapstructure:"tracing"`
	WebSocket     WebSocketConfig     `mapstructure:"websocket"`
	Etcd          EtcdConfig          `mapstructure:"etcd"`
}

// 应用配置
type AppConfig struct {
	Name    string `mapstructure:"name"`
	Env     string `mapstructure:"env"`
	Version string `mapstructure:"version"`
}

// 端口配置
type PortsConfig struct {
	Gateway     int `mapstructure:"gateway"`
	User        int `mapstructure:"user"`
	Video       int `mapstructure:"video"`
	Social      int `mapstructure:"social"`
	Interaction int `mapstructure:"interaction"`
	Message     int `mapstructure:"message"`
	Live        int `mapstructure:"live"`
	Danmu       int `mapstructure:"danmu"`
	Recommend   int `mapstructure:"recommend"`
	WebSocket   int `mapstructure:"websocket"`
}

// 服务配置
type ServicesConfig struct {
	User        ServiceConfig          `mapstructure:"user"`
	Video       ServiceConfig          `mapstructure:"video"`
	Social      ServiceConfig          `mapstructure:"social"`
	Interaction ServiceConfig          `mapstructure:"interaction"`
	Message     ServiceConfig          `mapstructure:"message"`
	Live        ServiceConfig          `mapstructure:"live"`
	Danmu       ServiceConfig          `mapstructure:"danmu"`
	Recommend   ServiceConfig          `mapstructure:"recommend"`
	WebSocket   WebSocketServiceConfig `mapstructure:"websocket"`
}

type ServiceConfig struct {
	Timeout string `mapstructure:"timeout"`
}

// WebSocket服务配置
type WebSocketServiceConfig struct {
	Timeout        string `mapstructure:"timeout"`
	MaxConnections int    `mapstructure:"max_connections"`
	PingInterval   string `mapstructure:"ping_interval"`
	WriteWait      string `mapstructure:"write_wait"`
	PongWait       string `mapstructure:"pong_wait"`
}

// WebSocket配置
type WebSocketConfig struct {
	Enable       bool     `mapstructure:"enable"`
	Path         string   `mapstructure:"path"`
	AllowOrigins []string `mapstructure:"allow_origins"`
	BufferSize   int      `mapstructure:"buffer_size"`
}

// 数据库配置
type DatabaseConfig struct {
	Postgres PostgresConfig `mapstructure:"postgres"`
}

type PostgresConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// Kafka配置
type KafkaConfig struct {
	Brokers []string     `mapstructure:"brokers"`
	Version string       `mapstructure:"version"`
	Topics  TopicsConfig `mapstructure:"topics"`
}

type TopicsConfig struct {
	User        string `mapstructure:"user"`
	Video       string `mapstructure:"video"`
	Interaction string `mapstructure:"interaction"`
	Social      string `mapstructure:"social"`
	Message     string `mapstructure:"message"`
	Live        string `mapstructure:"live"`
	Danmu       string `mapstructure:"danmu"`
	Recommend   string `mapstructure:"recommend"`
}

// Elasticsearch配置
type ElasticsearchConfig struct {
	URL      string `mapstructure:"url"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// MinIO配置
type MinioConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

// 日志配置
type LogConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

// JWT配置
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

// Prometheus配置
type PrometheusConfig struct {
	Enable          bool   `mapstructure:"enable"`
	Port            int    `mapstructure:"port"`
	Path            string `mapstructure:"path"`
	UserPort        int    `mapstructure:"user_port"`
	VideoPort       int    `mapstructure:"video_port"`
	SocialPort      int    `mapstructure:"social_port"`
	InteractionPort int    `mapstructure:"interaction_port"`
	MessagePort     int    `mapstructure:"message_port"`
	LivePort        int    `mapstructure:"live_port"`
	DanmuPort       int    `mapstructure:"danmu_port"`
	RecommendPort   int    `mapstructure:"recommend_port"`
	GatewayPort     int    `mapstructure:"gateway_port"`
}

// 链路追踪配置
type TracingConfig struct {
	Enable         bool    `mapstructure:"enable"`
	JaegerEndpoint string  `mapstructure:"jaeger_endpoint"`
	SampleRate     float64 `mapstructure:"sample_rate"`
}

// 加载配置
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")
	setDefaults()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("未找到配置文件，将使用默认值")
		} else {
			log.Printf("读取配置文件错误：%v", err)
			return nil, err
		}
	}
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

// 初始化配置
func Init() (*Config, error) {
	var err error
	configOnce.Do(func() {
		configInstance, err = LoadConfig()
	})
	return configInstance, err
}

// 获取配置实例
func Get() *Config {
	if configInstance == nil {
		panic("config not initialized")
	}
	return configInstance
}

// 设置默认值
func setDefaults() {
	viper.SetDefault("app.name", "shortvideo")
	viper.SetDefault("app.env", "dev")
	viper.SetDefault("app.version", "1.0.0")

	viper.SetDefault("ports.gateway", 8080)
	viper.SetDefault("ports.user", 8881)
	viper.SetDefault("ports.video", 8882)
	viper.SetDefault("ports.social", 8883)
	viper.SetDefault("ports.interaction", 8884)
	viper.SetDefault("ports.message", 8885)
	viper.SetDefault("ports.live", 8886)
	viper.SetDefault("ports.danmu", 8887)
	viper.SetDefault("ports.recommend", 8888)
	viper.SetDefault("ports.websocket", 8889)

	viper.SetDefault("services.user.timeout", "5s")
	viper.SetDefault("services.video.timeout", "5s")
	viper.SetDefault("services.social.timeout", "5s")
	viper.SetDefault("services.interaction.timeout", "5s")
	viper.SetDefault("services.message.timeout", "5s")
	viper.SetDefault("services.live.timeout", "5s")
	viper.SetDefault("services.danmu.timeout", "5s")
	viper.SetDefault("services.recommend.timeout", "5s")
	viper.SetDefault("services.websocket.timeout", "30s")
	viper.SetDefault("services.websocket.max_connections", 10000)
	viper.SetDefault("services.websocket.ping_interval", "30s")
	viper.SetDefault("services.websocket.write_wait", "10s")
	viper.SetDefault("services.websocket.pong_wait", "60s")

	viper.SetDefault("database.postgres.host", "localhost")
	viper.SetDefault("database.postgres.port", 5432)
	viper.SetDefault("database.postgres.user", "postgres")
	viper.SetDefault("database.postgres.password", "postgres")
	viper.SetDefault("database.postgres.dbname", "shortvideo")
	viper.SetDefault("database.postgres.sslmode", "disable")
	viper.SetDefault("database.postgres.max_open_conns", 100)
	viper.SetDefault("database.postgres.max_idle_conns", 10)
	viper.SetDefault("database.postgres.conn_max_lifetime", 3600)

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)

	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.version", "2.8.0")
	viper.SetDefault("kafka.topics.user", "user-events")
	viper.SetDefault("kafka.topics.video", "video-events")
	viper.SetDefault("kafka.topics.interaction", "interaction-events")
	viper.SetDefault("kafka.topics.social", "social-events")
	viper.SetDefault("kafka.topics.message", "message-events")
	viper.SetDefault("kafka.topics.live", "live-events")
	viper.SetDefault("kafka.topics.danmu", "danmu-events")
	viper.SetDefault("kafka.topics.recommend", "recommend-events")

	viper.SetDefault("elasticsearch.url", "http://localhost:9200")
	viper.SetDefault("elasticsearch.username", "")
	viper.SetDefault("elasticsearch.password", "")

	viper.SetDefault("minio.endpoint", "localhost:9000")
	viper.SetDefault("minio.access_key", "minioadmin")
	viper.SetDefault("minio.secret_key", "minioadmin")
	viper.SetDefault("minio.bucket", "shortvideo")
	viper.SetDefault("minio.use_ssl", false)

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "console")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("log.file_path", "./logs/app.log")

	viper.SetDefault("jwt.secret", "misonomika")
	viper.SetDefault("jwt.expire_hours", 168)

	viper.SetDefault("prometheus.enable", true)
	viper.SetDefault("prometheus.port", 9090)
	viper.SetDefault("prometheus.path", "/metrics")
	viper.SetDefault("prometheus.user_port", 9091)
	viper.SetDefault("prometheus.video_port", 9092)
	viper.SetDefault("prometheus.social_port", 9093)
	viper.SetDefault("prometheus.interaction_port", 9094)
	viper.SetDefault("prometheus.message_port", 9095)
	viper.SetDefault("prometheus.live_port", 9096)
	viper.SetDefault("prometheus.danmu_port", 9097)
	viper.SetDefault("prometheus.recommend_port", 9098)
	viper.SetDefault("prometheus.gateway_port", 9099)

	viper.SetDefault("tracing.enable", false)
	viper.SetDefault("tracing.jaeger_endpoint", "http://localhost:14268/api/traces")
	viper.SetDefault("tracing.sample_rate", 0.1)

	viper.SetDefault("websocket.enable", true)
	viper.SetDefault("websocket.path", "/ws")
	viper.SetDefault("websocket.allow_origins", []string{"*"})
	viper.SetDefault("websocket.buffer_size", 4096)

	viper.SetDefault("etcd.endpoints", []string{"localhost:2379"})
	viper.SetDefault("etcd.dial_timeout", "5s")
	viper.SetDefault("etcd.username", "")
	viper.SetDefault("etcd.password", "")
	viper.SetDefault("etcd.ttl", 30)
	viper.SetDefault("etcd.enable_secure", false)
}

// 获取PostgreSQL连接字符串
func (p *PostgresConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DBName, p.SSLMode)
}

// Etcd配置
type EtcdConfig struct {
	Endpoints    []string `mapstructure:"endpoints"`
	DialTimeout  string   `mapstructure:"dial_timeout"`
	Username     string   `mapstructure:"username"`
	Password     string   `mapstructure:"password"`
	TTL          int      `mapstructure:"ttl"`
	EnableSecure bool     `mapstructure:"enable_secure"`
}
