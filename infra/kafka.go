package infra

import (
	"fmt"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/segmentio/kafka-go"
)

func NewKafkaWriter(cfg config.KafkaConfig, topic string) *kafka.Writer {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	writer := &kafka.Writer{
		Addr:     kafka.TCP(addr),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return writer
}

func NewKafkaReader(cfg config.KafkaConfig, topic, groupId string) *kafka.Reader {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	readerCfg := kafka.ReaderConfig{
		Brokers:  []string{addr},
		Topic:    topic,
		GroupID:  groupId,
		MinBytes: 10 * 1024,        // 10KB
		MaxBytes: 10 * 1024 * 1024, // 10MB
	}
	reader := kafka.NewReader(readerCfg)
	return reader
}
