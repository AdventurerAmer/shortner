package brokers

import (
	"context"

	"github.com/AdventurerAmer/shortner/internal/core/ports"
)

type kafkaProducer struct {
}

func (p *kafkaProducer) Send(ctx context.Context, key string, data []byte) error {
	return nil
}

func NewKafkaProducer() ports.Producer {
	return &kafkaProducer{}
}
