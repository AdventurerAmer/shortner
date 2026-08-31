package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

func NewKafkaWriter(cfg config.KafkaConfig, topic domain.Topic) *kafka.Writer {
	mechanism := plain.Mechanism{
		Username: cfg.Username,
		Password: cfg.Password,
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
		Username: cfg.Username,
		Password: cfg.Password,
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

func PingKafka(ctx context.Context, cfg config.KafkaConfig) error {
	mechanism := plain.Mechanism{
		Username: cfg.Username,
		Password: cfg.Password,
	}

	dialer := &kafka.Dialer{
		DualStack:     true,
		SASLMechanism: mechanism,
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to broker: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			logger := logging.Get(ctx)
			logger.Error("'conn.Close' from PingKafka failed", "error", err)
		}
	}()

	if _, err := conn.ReadPartitions(); err != nil {
		return fmt.Errorf("failed to fetch metadata from kafka: %w", err)
	}

	return nil
}
