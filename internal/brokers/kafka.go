package brokers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
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
	logger := logging.Get(ctx)

loop:
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				break loop
			}
			logger.Error("'reader.FetchMessage' failed", "error", err)
			continue
		}

		logger.Debug("read message successfully")

		key := string(msg.Key)
		data := msg.Value
		if err := handler(ctx, key, data); err != nil {
			logger.Error("failed to handle message", "error", err)
			continue
		}

		logger.Debug("handled message successfully")

		var lastErr error
		for range 10 {
			err := func() error {
				dctx, cancel := context.WithTimeout(ctx, time.Second)
				defer cancel()

				if err := c.reader.CommitMessages(dctx, msg); err != nil {
					return fmt.Errorf("'reader.CommitMessages' failed: %w", err)
				}
				return nil
			}()
			if err != nil {
				lastErr = err
			} else {
				lastErr = nil
				break
			}
		}
		if lastErr != nil {
			logger.Error("ack message failed", "error", lastErr)
		} else {
			logger.Debug("acked message successfully")
		}
	}
	return nil
}
