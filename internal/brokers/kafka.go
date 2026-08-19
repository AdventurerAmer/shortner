package brokers

import (
	"context"
	"errors"
	"fmt"

	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/telemetry"
	"github.com/segmentio/kafka-go"
)

type kafkaProducer struct {
	writer *kafka.Writer
}

func (p *kafkaProducer) Send(ctx context.Context, key string, data []byte) error {
	dctx, span := telemetry.NewSpan(ctx, "kafkaProducer: Send")
	defer span.End()

	msg := kafka.Message{
		Key:   []byte(key),
		Value: data,
	}
	if err := p.writer.WriteMessages(dctx, msg); err != nil {
		span.RecordError(err)
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

func (c *kafkaConsumer) Receive(ctx context.Context) <-chan ports.ConsumerMessage {
	dctx, span := telemetry.NewSpan(ctx, "kafkaConsumer: Receive")
	defer span.End()

	logger := logging.Get(ctx)

	msgCh := make(chan ports.ConsumerMessage, 1)
	doneCh := make(chan struct{})

	go func() {
		for {
			msg, err := c.reader.FetchMessage(dctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					span.RecordError(err)
					close(doneCh)
					close(msgCh)
					return
				}
				logger.Error("'reader.FetchMessage' failed", "error", err)
				continue
			}
			key := string(msg.Key)
			data := msg.Value
			msgCh <- ports.ConsumerMessage{
				Key:         key,
				Data:        data,
				OriginalMsg: msg,
			}
		}
	}()

	return msgCh
}

func (c *kafkaConsumer) Ack(ctx context.Context, msg ports.ConsumerMessage) error {
	dctx, span := telemetry.NewSpan(ctx, "kafkaConsumer: Ack")
	defer span.End()

	originalMsg := msg.OriginalMsg.(kafka.Message)
	if err := c.reader.CommitMessages(dctx, originalMsg); err != nil {
		span.RecordError(err)
		return fmt.Errorf("'reader.CommitMessages' failed: %w", err)
	}
	return nil
}
