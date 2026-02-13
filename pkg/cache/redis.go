package cache

import (
	"context"
	"fmt"
	"shortvideo/pkg/config"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache interface {
	//基本操作
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	//批量操作
	MSet(ctx context.Context, values map[string]interface{}) error
	MGet(ctx context.Context, keys []string) (map[string]string, error)
	MDelete(ctx context.Context, keys []string) error

	//哈希操作
	HSet(ctx context.Context, key, field string, value interface{}) error
	HGet(ctx context.Context, key, field string) (string, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, fields []string) error

	//列表操作
	LPush(ctx context.Context, key string, values ...interface{}) error
	RPush(ctx context.Context, key string, values ...interface{}) error
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	LLen(ctx context.Context, key string) (int64, error)

	//集合操作
	SAdd(ctx context.Context, key string, members ...interface{}) error
	SRem(ctx context.Context, key string, members ...interface{}) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SIsMember(ctx context.Context, key string, member interface{}) (bool, error)

	//有序集合操作
	ZAdd(ctx context.Context, key string, score float64, member interface{}) error
	ZRem(ctx context.Context, key string, members ...interface{}) error
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZScore(ctx context.Context, key string, member string) (float64, error)

	//计数器操作
	Incr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	DecrBy(ctx context.Context, key string, value int64) (int64, error)

	//连接管理
	Ping(ctx context.Context) error
	Close() error
}

type RedisCache struct {
	client *redis.Client
}

var (
	cacheInstance Cache
	cacheOnce     sync.Once
)

func NewRedisCache() Cache {
	cacheOnce.Do(func() {
		cfg := config.Get()
		cacheInstance, _ = InitRedis(cfg.Redis)
	})
	return cacheInstance
}

func InitRedis(redisConfig config.RedisConfig) (Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
		PoolSize: redisConfig.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("Redis连接失败: %v\n", err)
		return nil, err
	} else {
		fmt.Println("Redis连接成功")
	}

	return &RedisCache{client: client}, nil
}



// 设置键值对
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// 获取值
func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// 删除键
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// 检查键是否存在
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// 批量设置
func (c *RedisCache) MSet(ctx context.Context, values map[string]interface{}) error {
	return c.client.MSet(ctx, values).Err()
}

// 批量获取
func (c *RedisCache) MGet(ctx context.Context, keys []string) (map[string]string, error) {
	result, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	values := make(map[string]string)
	for i, key := range keys {
		if result[i] != nil {
			values[key] = result[i].(string)
		}
	}
	return values, nil
}

// 批量删除
func (c *RedisCache) MDelete(ctx context.Context, keys []string) error {
	return c.client.Del(ctx, keys...).Err()
}

// 设置哈希字段
func (c *RedisCache) HSet(ctx context.Context, key, field string, value interface{}) error {
	return c.client.HSet(ctx, key, field, value).Err()
}

// 获取哈希字段
func (c *RedisCache) HGet(ctx context.Context, key, field string) (string, error) {
	return c.client.HGet(ctx, key, field).Result()
}

// 获取哈希所有字段
func (c *RedisCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

// 删除哈希字段
func (c *RedisCache) HDel(ctx context.Context, key string, fields []string) error {
	return c.client.HDel(ctx, key, fields...).Err()
}

// 左侧推入列表
func (c *RedisCache) LPush(ctx context.Context, key string, values ...interface{}) error {
	return c.client.LPush(ctx, key, values...).Err()
}

// 右侧推入列表
func (c *RedisCache) RPush(ctx context.Context, key string, values ...interface{}) error {
	return c.client.RPush(ctx, key, values...).Err()
}

// 获取列表范围
func (c *RedisCache) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.LRange(ctx, key, start, stop).Result()
}

// 获取列表长度
func (c *RedisCache) LLen(ctx context.Context, key string) (int64, error) {
	return c.client.LLen(ctx, key).Result()
}

// 添加集合成员
func (c *RedisCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SAdd(ctx, key, members...).Err()
}

// 删除集合成员
func (c *RedisCache) SRem(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SRem(ctx, key, members...).Err()
}

// 获取集合所有成员
func (c *RedisCache) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, key).Result()
}

// 检查是否为集合成员
func (c *RedisCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.client.SIsMember(ctx, key, member).Result()
}

// 添加有序集合成员
func (c *RedisCache) ZAdd(ctx context.Context, key string, score float64, member interface{}) error {
	return c.client.ZAdd(ctx, key, &redis.Z{Score: score, Member: member}).Err()
}

// 删除有序集合成员
func (c *RedisCache) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return c.client.ZRem(ctx, key, members...).Err()
}

// 获取有序集合范围
func (c *RedisCache) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.ZRange(ctx, key, start, stop).Result()
}

// 获取有序集合成员分数
func (c *RedisCache) ZScore(ctx context.Context, key string, member string) (float64, error) {
	return c.client.ZScore(ctx, key, member).Result()
}

// 递增计数器
func (c *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// 递增指定值
func (c *RedisCache) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

// 递减计数器
func (c *RedisCache) Decr(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

// 递减指定值
func (c *RedisCache) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.DecrBy(ctx, key, value).Result()
}

// 测试连接
func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// 关闭连接
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// 生成用户缓存键
func GenerateUserKey(userID int64) string {
	return fmt.Sprintf("user:%d", userID)
}

// 生成视频缓存键
func GenerateVideoKey(videoID int64) string {
	return fmt.Sprintf("video:%d", videoID)
}

// 生成用户视频列表缓存键
func GenerateUserVideosKey(userID int64) string {
	return fmt.Sprintf("user:videos:%d", userID)
}

// 生成视频流缓存键
func GenerateFeedKey(userID int64) string {
	return fmt.Sprintf("feed:%d", userID)
}
