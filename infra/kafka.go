package infra

import (
	"fmt"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/segmentio/kafka-go"
)

func NewKafkaWriter(cfg config.KafkaConfig, topic domain.Topic) *kafka.Writer {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(addr),
		Topic:                  string(topic),
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireOne,
		Async:                  true,
		AllowAutoTopicCreation: true,
	}
	return writer
}

func NewKafkaReader(cfg config.KafkaConfig, topic domain.Topic, groupId string) *kafka.Reader {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	readerCfg := kafka.ReaderConfig{
		Brokers:     []string{addr},
		Topic:       string(topic),
		GroupID:     groupId,
		StartOffset: kafka.FirstOffset,
		MinBytes:    10 * 1024,        // 10KB
		MaxBytes:    10 * 1024 * 1024, // 10MB
	}
	reader := kafka.NewReader(readerCfg)
	return reader
}
