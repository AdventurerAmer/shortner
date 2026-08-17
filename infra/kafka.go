package infra

import (
	"fmt"
	"time"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

func NewKafkaWriter(cfg config.KafkaConfig, topic domain.Topic) *kafka.Writer {
	mechanism := plain.Mechanism{
		Username: "admin",
		Password: "admin",
	}

	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: mechanism,
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	writerCfg := kafka.WriterConfig{
		Brokers:      []string{addr},
		Topic:        string(topic),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: 1,
		Async:        false,
		Dialer:       dialer,
	}
	writer := kafka.NewWriter(writerCfg)
	return writer
}

func NewKafkaReader(cfg config.KafkaConfig, topic domain.Topic, groupId string) *kafka.Reader {
	mechanism := plain.Mechanism{
		Username: "admin",
		Password: "admin",
	}

	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: mechanism,
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	readerCfg := kafka.ReaderConfig{
		Brokers:     []string{addr},
		Topic:       string(topic),
		GroupID:     groupId,
		StartOffset: kafka.LastOffset,
		MinBytes:    10 * 1024,        // 10KB
		MaxBytes:    10 * 1024 * 1024, // 10MB
		Dialer:      dialer,
	}
	reader := kafka.NewReader(readerCfg)
	return reader
}
