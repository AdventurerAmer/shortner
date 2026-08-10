package brokers

import (
	"context"
	"errors"
	"fmt"

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

func (c *kafkaConsumer) Receive(ctx context.Context) <-chan ports.ConsumerMessage {
	logger := logging.Get(ctx)

	msgCh := make(chan ports.ConsumerMessage, 1)
	doneCh := make(chan struct{})

	go func() {
		for {
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
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
	originalMsg := msg.OriginalMsg.(kafka.Message)
	if err := c.reader.CommitMessages(ctx, originalMsg); err != nil {
		return fmt.Errorf("'reader.CommitMessages' failed: %w", err)
	}
	return nil
}
