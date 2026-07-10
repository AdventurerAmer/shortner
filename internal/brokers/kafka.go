package brokers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/segmentio/kafka-go"
)

type kafkaProducer struct {
	writer *kafka.Writer
}

func (p *kafkaProducer) Send(ctx context.Context, key string, data []byte) error {
	msg := kafka.Message{
		Key:   []byte(key),
		Value: data,
	}
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("'writer.WriteMessages' failed: %w", err)
	}
	return nil
}

func NewKafkaProducer(writer *kafka.Writer) ports.Producer {
	return &kafkaProducer{
		writer: writer,
	}
}

type kafkaConsumer struct {
	reader *kafka.Reader
}

func NewKafkaConsumer(reader *kafka.Reader) ports.Consumer {
	return &kafkaConsumer{
		reader: reader,
	}
}

func (c *kafkaConsumer) Receive(ctx context.Context, handler ports.ConsumerHandlerFunc) error {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				break
			}
			slog.Error("'reader.ReadMessage' failed", "error", err)
			continue
		}
		key := string(msg.Key)
		data := msg.Value
		handler(key, data)
	}
	return nil
}
