package brokers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/telemetry"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
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

	carrier := KafkaGoCarrier(msg.Headers)
	otel.GetTextMapPropagator().Inject(ctx, &carrier)
	msg.Headers = []kafka.Header(carrier)

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

func (c *kafkaConsumer) Consume(ctx context.Context, msg ports.ConsumerMessage, handler ports.ConsumerHandlerFunc) error {
	originalMsg := msg.OriginalMsg.(kafka.Message)

	dctx, span := NewConsumerSpan(ctx, "kafkaConsumer: Consume", originalMsg)
	defer span.End()

	if err := handler(dctx, msg); err != nil {
		return fmt.Errorf("'handler' failed: %w", err)
	}

	return nil
}

func (c *kafkaConsumer) Ack(ctx context.Context, msg ports.ConsumerMessage) error {
	originalMsg := msg.OriginalMsg.(kafka.Message)

	dctx, span := NewConsumerSpan(ctx, "kafkaConsumer: Ack", originalMsg)
	defer span.End()

	if err := c.reader.CommitMessages(dctx, originalMsg); err != nil {
		span.RecordError(err)
		return fmt.Errorf("'reader.CommitMessages' failed: %w", err)
	}
	return nil
}

func NewConsumerSpan(ctx context.Context, name string, msg kafka.Message) (context.Context, telemetry.Span) {
	// 1. Extract remote context from headers
	carrier := KafkaGoCarrier(msg.Headers)
	dctx := otel.GetTextMapPropagator().Extract(ctx, &carrier)
	tr := telemetry.GetTracer()
	// Option A – child span (keeps parent-child relationship)
	tctx, span := tr.Start(dctx, "kafka.consume",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			semconv.MessagingSystemKey.String("kafka"),
			semconv.MessagingDestinationName(msg.Topic),
			semconv.MessagingOperationReceive,
			attribute.Int("messaging.kafka.partition", msg.Partition),
			attribute.Int64("messaging.kafka.offset", msg.Offset),
		),
	)

	traceId := telemetry.GetTraceId(tctx)
	logger := logging.Get(tctx).With(slog.String("correlationId", traceId))
	lctx := logging.Set(tctx, logger)

	return lctx, span
}

// KafkaGoCarrier adapts []kafka.Header so OpenTelemetry can inject/extract
// trace context (traceparent, tracestate, baggage) into Kafka message headers.
type KafkaGoCarrier []kafka.Header

func (c *KafkaGoCarrier) Get(key string) string {
	for _, h := range *c {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c *KafkaGoCarrier) Set(key, value string) {
	for i, h := range *c {
		if h.Key == key {
			(*c)[i].Value = []byte(value)
			return
		}
	}
	*c = append(*c, kafka.Header{Key: key, Value: []byte(value)})
}

func (c *KafkaGoCarrier) Keys() []string {
	keys := make([]string, len(*c))
	for i, h := range *c {
		keys[i] = h.Key
	}
	return keys
}

// Ensure it satisfies the interface at compile time
var _ propagation.TextMapCarrier = (*KafkaGoCarrier)(nil)
