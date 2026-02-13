package mq

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"shortvideo/pkg/config"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

type Consumer struct {
	reader *kafka.Reader
}

type Message struct {
	Topic     string
	Key       string
	Value     []byte
	Partition int
	Offset    int64
	Time      time.Time
}

var (
	producerInstance *Producer
	producerOnce     sync.Once
)

func NewProducer() *Producer {
	producerOnce.Do(func() {
		kafkaConfig := config.Get().Kafka
		producerInstance, _ = InitProducer(kafkaConfig.Brokers, kafkaConfig.Version)
		ctx := context.Background()
		if err := CreateTopics(ctx); err != nil {
			log.Printf("创建Kafka主题失败: %v", err)
		}
	})
	return producerInstance
}

func InitProducer(brokers []string, version string) (*Producer, error) {
	producer := &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Balancer:     &kafka.LeastBytes{},
			WriteTimeout: 10 * time.Second,
			ReadTimeout:  10 * time.Second,
		},
	}
	return producer, nil
}

func NewConsumer(topic string, groupID string) *Consumer {
	kafkaConfig := config.Get().Kafka
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:         kafkaConfig.Brokers,
			Topic:           topic,
			GroupID:         groupID,
			MinBytes:        10e3,
			MaxBytes:        10e6,
			MaxWait:         10 * time.Second,
			ReadLagInterval: time.Second,
		}),
	}
}

func (p *Producer) Send(ctx context.Context, topic string, key string, value []byte) error {
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
	}

	err := p.writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

	log.Printf("消息已发送到主题 %s: key=%s, value=%s", topic, key, string(value))
	return nil
}

func (p *Producer) SendUserEvent(ctx context.Context, key string, value []byte) error {
	topic := config.Get().Kafka.Topics.User
	return p.Send(ctx, topic, key, value)
}

func (p *Producer) SendVideoEvent(ctx context.Context, key string, value []byte) error {
	topic := config.Get().Kafka.Topics.Video
	return p.Send(ctx, topic, key, value)
}

func (p *Producer) SendInteractionEvent(ctx context.Context, key string, value []byte) error {
	topic := config.Get().Kafka.Topics.Interaction
	return p.Send(ctx, topic, key, value)
}

func (p *Producer) SendSocialEvent(ctx context.Context, key string, value []byte) error {
	topic := config.Get().Kafka.Topics.Social
	return p.Send(ctx, topic, key, value)
}

func (p *Producer) SendMessageEvent(ctx context.Context, key string, value []byte) error {
	topic := config.Get().Kafka.Topics.Message
	return p.Send(ctx, topic, key, value)
}

func (p *Producer) SendLiveEvent(ctx context.Context, key string, value []byte) error {
	topic := config.Get().Kafka.Topics.Live
	return p.Send(ctx, topic, key, value)
}

func (p *Producer) SendDanmuEvent(ctx context.Context, key string, value []byte) error {
	topic := config.Get().Kafka.Topics.Danmu
	return p.Send(ctx, topic, key, value)
}

func (p *Producer) SendRecommendEvent(ctx context.Context, key string, value []byte) error {
	topic := config.Get().Kafka.Topics.Recommend
	return p.Send(ctx, topic, key, value)
}

func (c *Consumer) Receive(ctx context.Context) (*Message, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("读取消息失败: %w", err)
	}

	return &Message{
		Topic:     msg.Topic,
		Key:       string(msg.Key),
		Value:     msg.Value,
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Time:      msg.Time,
	}, nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func CreateTopics(ctx context.Context) error {
	kafkaConfig := config.Get().Kafka
	log.Printf("Kafka主题配置: %v", kafkaConfig.Topics)
	log.Printf("Kafka brokers: %v", kafkaConfig.Brokers)

	if len(kafkaConfig.Brokers) == 0 {
		log.Println("Kafka brokers配置为空")
		return nil
	}

	conn, err := kafka.Dial("tcp", kafkaConfig.Brokers[0])
	if err != nil {
		log.Printf("连接Kafka失败: %v", err)
		return err
	}
	defer conn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             kafkaConfig.Topics.User,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             kafkaConfig.Topics.Video,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             kafkaConfig.Topics.Interaction,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             kafkaConfig.Topics.Social,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             kafkaConfig.Topics.Message,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             kafkaConfig.Topics.Live,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             kafkaConfig.Topics.Danmu,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             kafkaConfig.Topics.Recommend,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	for _, tc := range topicConfigs {
		log.Printf("尝试创建主题: %s", tc.Topic)
		err := conn.CreateTopics(tc)
		if err != nil {
			log.Printf("创建主题 %s 失败: %v", tc.Topic, err)
			continue
		}
		log.Printf("主题 %s 创建成功", tc.Topic)
	}

	return nil
}
